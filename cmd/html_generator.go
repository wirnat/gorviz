package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strings"

	"github.com/wirnat/gorviz/internal"
)

type templateData struct {
	Models     []internal.Model
	SchemaJSON template.JS
}

func generateStaticHTML(schema *internal.Schema) (string, error) {
	// Sort models alphabetically by name
	sort.Slice(schema.Models, func(i, j int) bool {
		return schema.Models[i].Name < schema.Models[j].Name
	})

	// Serialize schema to JSON for JS usage
	jsonData, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("doc").Funcs(template.FuncMap{
		"formatTags": func(tags map[string]string) string {
			var parts []string
			for k, v := range tags {
				parts = append(parts, fmt.Sprintf("%s: %s", k, v))
			}
			return strings.Join(parts, " | ")
		},
		"formatType": func(t string) string {
			return strings.TrimPrefix(t, "*")
		},
	}).Parse(docTemplate)

	if err != nil {
		return "", err
	}

	data := templateData{
		Models:     schema.Models,
		SchemaJSON: template.JS(jsonData),
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}

	return b.String(), nil
}
