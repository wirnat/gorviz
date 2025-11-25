package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/wirnat/gorviz/internal"
)

// ParseGormModels scans a directory for Go files, extracts GORM model definitions,
// and returns them as a Schema object.
func ParseGormModels(dirPath string) (*internal.Schema, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dirPath, func(info os.FileInfo) bool {
		return !info.IsDir() && strings.HasSuffix(info.Name(), ".go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	schema := &internal.Schema{
		Models: []internal.Model{},
	}

	structDefinitions := make(map[string]*ast.StructType)
	structFiles := make(map[string]*ast.File)
	hasTableNameMethod := make(map[string]bool)
	customTableNames := make(map[string]string)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				// Capture Struct Definitions
				if typeSpec, ok := node.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						structName := typeSpec.Name.Name
						structDefinitions[structName] = structType
						structFiles[structName] = file
					}
				}

				// Capture TableName() methods to identify actual GORM models
				if funcDecl, ok := node.(*ast.FuncDecl); ok {
					if funcDecl.Name.Name == "TableName" && funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
						// Get receiver type name
						recvType := funcDecl.Recv.List[0].Type
						var structName string
						
						// Handle pointer receivers (*User) and value receivers (User)
						if starExpr, ok := recvType.(*ast.StarExpr); ok {
							if ident, ok := starExpr.X.(*ast.Ident); ok {
								structName = ident.Name
							}
						} else if ident, ok := recvType.(*ast.Ident); ok {
							structName = ident.Name
						}
						
						if structName != "" {
							hasTableNameMethod[structName] = true

							// Try to extract simple return "string" for table name
							if funcDecl.Body != nil && len(funcDecl.Body.List) > 0 {
								if retStmt, ok := funcDecl.Body.List[0].(*ast.ReturnStmt); ok && len(retStmt.Results) > 0 {
									if basicLit, ok := retStmt.Results[0].(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
										customTableNames[structName] = strings.Trim(basicLit.Value, `"`)
									}
								}
							}
						}
					}
				}
				return true
			})
		}
	}

	for structName, structType := range structDefinitions {
		// FILTER: Only treat as a Table/Model if it has a TableName() method
		if !hasTableNameMethod[structName] {
			continue
		}

		model := internal.Model{
			Name:          structName,
			Fields:        []internal.Field{},
			Relationships: []internal.Relationship{},
		}

		// Determine Table Name: Custom Method > GORM Tag > Convention
		if customName, ok := customTableNames[structName]; ok {
			model.TableName = customName
		} else {
			model.TableName = toTableName(structName)
			if gormTag, ok := extractGormTag(structType.Fields); ok {
				parsedGormTags := parseGormTag(gormTag)
				if tableName, ok := parsedGormTags["table_name"]; ok {
					model.TableName = strings.Trim(tableName, `"`)
				}
			}
		}

		for _, field := range structType.Fields.List {
			fieldTypeStr := getFieldType(field.Type)

			if len(field.Names) == 0 { // This is an embedded field
				embeddedFieldName := getEmbeddedFieldName(field.Type)
				isExternalEmbedded := isSelectorExprType(field.Type)
				
				// NOTE: We still process embedded structs even if they don't have TableName()
				// because they are part of the parent table's structure.
				if embeddedStruct, ok := structDefinitions[embeddedFieldName]; ok && !isExternalEmbedded {
					visited := map[string]bool{structName: true}
					embeddedModelFields, embeddedRelationships, err := processEmbeddedStruct(structName, embeddedFieldName, embeddedStruct, structDefinitions, structFiles, visited)
					if err != nil {
						return nil, fmt.Errorf("failed to process embedded struct %s in %s: %w", embeddedFieldName, structName, err)
					}
					model.Fields = append(model.Fields, embeddedModelFields...)
					model.Relationships = append(model.Relationships, embeddedRelationships...)
					model.Relationships = append(model.Relationships, internal.Relationship{
						SourceModelName: structName,
						TargetModelName: embeddedFieldName,
						Type:            internal.Embedded,
						FieldName:       embeddedFieldName,
					})
				} else {
					// fmt.Printf("Warning: Embedded struct '%s' in '%s' not found in parsed files or external. Using '%s' as field name.\n", fieldTypeStr, structName, embeddedFieldName)
					model.Fields = append(model.Fields, internal.Field{
						Name:       embeddedFieldName,
						Type:       fieldTypeStr,
						IsEmbedded: true,
					})
				}
				continue
			}

			// Regular field processing
			fieldItem := internal.Field{
				Name: field.Names[0].Name,
				Type: fieldTypeStr,
				Tags: parseStructTag(field.Tag),
			}

			gormTagValue := ""
			if tags, ok := fieldItem.Tags["gorm"]; ok {
				gormTagValue = strings.Trim(tags, `"`)
			}
			parsedGormTags := parseGormTag(gormTagValue)

			if _, ok := parsedGormTags["primarykey"]; ok || strings.ToLower(fieldItem.Name) == "id" {
				fieldItem.IsPrimaryKey = true
			}

			relType, relTargetModel, relForeignKey, relReferences, relJoinTable, relThroughModel := extractRelationshipFromGormTag(structName, fieldItem.Name, fieldTypeStr, parsedGormTags)
			if relType != internal.UnknownRel {
				model.Relationships = append(model.Relationships, internal.Relationship{
					SourceModelName: structName,
					TargetModelName: relTargetModel,
					Type:            relType,
					ForeignKey:      relForeignKey,
					References:      relReferences,
					JoinTable:       relJoinTable,
					ThroughModel:    relThroughModel,
					FieldName:       fieldItem.Name,
				})
				if relType == internal.BelongsTo || strings.Contains(gormTagValue, "foreignKey") {
					fieldItem.IsForeignKey = true
				}
			} else if _, ok := structDefinitions[fieldTypeStr]; ok && fieldTypeStr != structName {
				// Implicit BelongsTo check
				if !strings.HasPrefix(fieldTypeStr, "[]") &&
					strings.HasSuffix(fieldItem.Name, "ID") &&
					strings.HasPrefix(fieldItem.Name, fieldTypeStr) {
					model.Relationships = append(model.Relationships, internal.Relationship{
						SourceModelName: structName,
						TargetModelName: fieldTypeStr,
						Type:            internal.BelongsTo,
						ForeignKey:      fieldItem.Name,
						References:      "ID",
						FieldName:       fieldItem.Name,
					})
					fieldItem.IsForeignKey = true
				}
			}
			model.Fields = append(model.Fields, fieldItem)
		}
		schema.Models = append(schema.Models, model)
	}

	return schema, nil
}

func processEmbeddedStruct(
	parentStructName string,
	embeddedStructName string,
	embeddedStructType *ast.StructType,
	structDefinitions map[string]*ast.StructType,
	structFiles map[string]*ast.File,
	visited map[string]bool,
) ([]internal.Field, []internal.Relationship, error) {

	fields := []internal.Field{}
	relationships := []internal.Relationship{}

	if visited == nil {
		visited = make(map[string]bool)
	}
	if visited[embeddedStructName] {
		fmt.Printf("Warning: Detected circular embedded struct reference to '%s' while processing '%s'. Skipping to avoid infinite recursion.\n", embeddedStructName, parentStructName)
		return nil, nil, nil
	}
	visited[embeddedStructName] = true
	defer delete(visited, embeddedStructName)

	for _, field := range embeddedStructType.Fields.List {
		fieldTypeStr := getFieldType(field.Type)

		if len(field.Names) == 0 { // Nested embedded struct
			nestedEmbeddedStructName := getEmbeddedFieldName(field.Type)
			isExternalEmbedded := isSelectorExprType(field.Type)
			if nestedEmbeddedStruct, ok := structDefinitions[nestedEmbeddedStructName]; ok && !isExternalEmbedded {
				nestedFields, nestedRelationships, err := processEmbeddedStruct(parentStructName, nestedEmbeddedStructName, nestedEmbeddedStruct, structDefinitions, structFiles, visited)
				if err != nil {
					return nil, nil, err
				}
				fields = append(fields, nestedFields...)
				relationships = append(relationships, nestedRelationships...)
				relationships = append(relationships, internal.Relationship{
					SourceModelName: parentStructName,
					TargetModelName: nestedEmbeddedStructName,
					Type:            internal.Embedded,
					FieldName:       nestedEmbeddedStructName,
				})
			} else {
				fmt.Printf("Warning: Nested embedded struct '%s' in '%s' (embedded in '%s') not found. Using '%s' as field name.\n", fieldTypeStr, embeddedStructName, parentStructName, nestedEmbeddedStructName)
				fields = append(fields, internal.Field{
					Name:       nestedEmbeddedStructName,
					Type:       fieldTypeStr,
					IsEmbedded: true,
				})
			}
			continue
		}

		fieldItem := internal.Field{
			Name:       field.Names[0].Name,
			Type:       fieldTypeStr,
			Tags:       parseStructTag(field.Tag),
			IsEmbedded: true,
		}

		gormTagValue := ""
		if tags, ok := fieldItem.Tags["gorm"]; ok {
			gormTagValue = strings.Trim(tags, `"`)
		}
		parsedGormTags := parseGormTag(gormTagValue)

		if _, ok := parsedGormTags["primarykey"]; ok || (fieldItem.Name == "ID" && fieldItem.IsEmbedded) {
			fieldItem.IsPrimaryKey = true
		}

		relType, relTargetModel, relForeignKey, relReferences, relJoinTable, relThroughModel := extractRelationshipFromGormTag(parentStructName, fieldItem.Name, fieldTypeStr, parsedGormTags)
		if relType != internal.UnknownRel {
			relationships = append(relationships, internal.Relationship{
				SourceModelName: parentStructName,
				TargetModelName: relTargetModel,
				Type:            relType,
				ForeignKey:      relForeignKey,
				References:      relReferences,
				JoinTable:       relJoinTable,
				ThroughModel:    relThroughModel,
				FieldName:       fieldItem.Name,
			})
			if relType == internal.BelongsTo || strings.Contains(gormTagValue, "foreignKey") {
				fieldItem.IsForeignKey = true
			}
		} else if _, ok := structDefinitions[fieldTypeStr]; ok && fieldTypeStr != parentStructName {
			if !strings.HasPrefix(fieldTypeStr, "[]") &&
				strings.HasSuffix(fieldItem.Name, "ID") &&
				strings.HasPrefix(fieldItem.Name, fieldTypeStr) {
				relationships = append(relationships, internal.Relationship{
					SourceModelName: parentStructName,
					TargetModelName: fieldTypeStr,
					Type:            internal.BelongsTo,
					ForeignKey:      fieldItem.Name,
					References:      "ID",
					FieldName:       fieldItem.Name,
				})
				fieldItem.IsForeignKey = true
			}
		}

		fields = append(fields, fieldItem)
	}

	return fields, relationships, nil
}

func getFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", t.X, t.Sel.Name)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", getFieldType(t.X))
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", getFieldType(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getFieldType(t.Key), getFieldType(t.Value))
	default:
		return fmt.Sprintf("unknown_type_%T", t)
	}
}

func getEmbeddedFieldName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.StarExpr:
		return getEmbeddedFieldName(t.X)
	case *ast.ArrayType:
		return getEmbeddedFieldName(t.Elt)
	case *ast.MapType:
		return getEmbeddedFieldName(t.Value)
	default:
		return getFieldType(expr)
	}
}

func isSelectorExprType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		return true
	case *ast.StarExpr:
		return isSelectorExprType(t.X)
	case *ast.ArrayType:
		return isSelectorExprType(t.Elt)
	case *ast.MapType:
		return isSelectorExprType(t.Value)
	default:
		return false
	}
}

func parseStructTag(tag *ast.BasicLit) map[string]string {
	if tag == nil {
		return nil
	}
	tags := make(map[string]string)
	rawTag := strings.Trim(tag.Value, "`")

	parts := strings.FieldsFunc(rawTag, func(r rune) bool {
		return r == ' '
	})

	for _, part := range parts {
		if idx := strings.Index(part, ":"); idx != -1 {
			key := part[:idx]
			value := part[idx+1:]
			tags[key] = value
		}
	}
	return tags
}

func extractGormTag(fields *ast.FieldList) (string, bool) {
	if fields == nil {
		return "", false
	}
	for _, field := range fields.List {
		if field.Tag != nil {
			tags := parseStructTag(field.Tag)
			if gormTag, ok := tags["gorm"]; ok {
				return strings.Trim(gormTag, `"`), true
			}
		}
	}
	return "", false
}

func parseGormTag(gormTagValue string) map[string]string {
	options := make(map[string]string)
	if gormTagValue == "" {
		return options
	}

	parts := strings.Split(gormTagValue, ";")
	for _, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			options[kv[0]] = kv[1]
		} else {
			options[kv[0]] = ""
		}
	}
	return options
}

func extractRelationshipFromGormTag(
	sourceModelName, fieldName, fieldTypeStr string,
	parsedGormTags map[string]string,
) (internal.RelationshipType, string, string, string, string, string) {

	relType := internal.UnknownRel
	targetModelName := ""
	foreignKey := ""
	references := ""
	joinTable := ""
	throughModel := ""

	if strings.HasPrefix(fieldTypeStr, "[]") {
		targetModelName = strings.TrimPrefix(fieldTypeStr, "[]")
		relType = internal.HasMany
	} else if strings.HasPrefix(fieldTypeStr, "*") {
		targetModelName = strings.TrimPrefix(fieldTypeStr, "*")
	} else {
		targetModelName = fieldTypeStr
	}

	if jt, ok := parsedGormTags["many2many"]; ok {
		relType = internal.Many2Many
		joinTable = strings.Trim(jt, `"`)
		if tm, ok := parsedGormTags["through"]; ok {
			throughModel = strings.Trim(tm, `"`)
		}
	}

	if fk, ok := parsedGormTags["foreignKey"]; ok {
		foreignKey = strings.Trim(fk, `"`)
		if ref, ok := parsedGormTags["references"]; ok {
			references = strings.Trim(ref, `"`)
		}

		if strings.HasPrefix(fieldTypeStr, "[]") {
			relType = internal.HasMany
		} else {
			if relType == internal.UnknownRel {
				if foreignKey != "" {
					relType = internal.BelongsTo
				}
			}
		}
	}

	if relType == internal.UnknownRel && !strings.HasPrefix(fieldTypeStr, "[]") {
		if strings.HasSuffix(fieldName, "ID") && strings.HasPrefix(fieldName, targetModelName) {
			relType = internal.BelongsTo
			foreignKey = fieldName
			references = "ID"
		}
	}

	return relType, targetModelName, foreignKey, references, joinTable, throughModel
}

func toTableName(structName string) string {
	if structName == "" {
		return ""
	}

	runes := []rune(structName)
	var sb strings.Builder
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			sb.WriteRune('_')
		}
		sb.WriteRune(r)
	}
	snakeCase := strings.ToLower(sb.String())

	if !strings.HasSuffix(snakeCase, "s") {
		snakeCase += "s"
	}
	return snakeCase
}
