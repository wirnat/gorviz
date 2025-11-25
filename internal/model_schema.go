package internal

// Schema represents the top-level structure for the GORM visualization YAML.
type Schema struct {
	Models []Model `yaml:"models" json:"models"`
}

// Model represents a single GORM model (Go struct).
type Model struct {
	Name        string       `yaml:"name" json:"name"`
	TableName   string       `yaml:"table_name,omitempty" json:"table_name,omitempty"`
	Description string       `yaml:"description,omitempty" json:"description,omitempty"` // Add description for HTML display
	Fields      []Field      `yaml:"fields" json:"fields"`
	Relationships []Relationship `yaml:"relationships,omitempty" json:"relationships,omitempty"`
}

// Field represents a single field within a GORM model.
type Field struct {
	Name         string            `yaml:"name" json:"name"`
	Type         string            `yaml:"type" json:"type"`
	Tags         map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"` // Example: json, gorm
	IsPrimaryKey bool              `yaml:"is_primary_key,omitempty" json:"is_primary_key,omitempty"`
	IsForeignKey bool              `yaml:"is_foreign_key,omitempty" json:"is_foreign_key,omitempty"`
	IsEmbedded   bool              `yaml:"is_embedded,omitempty" json:"is_embedded,omitempty"` // To indicate if this field is an embedded struct itself
}

// RelationshipType defines the type of relationship between models.
type RelationshipType string

const (
	HasOne     RelationshipType = "has_one"
	HasMany    RelationshipType = "has_many"
	BelongsTo  RelationshipType = "belongs_to"
	Many2Many  RelationshipType = "many_to_many"
	Embedded   RelationshipType = "embedded" // Special type for embedded structs
	UnknownRel RelationshipType = "unknown"
)

// Relationship represents a relationship between two GORM models.
type Relationship struct {
	SourceModelName string           `yaml:"source_model_name" json:"source_model_name"`
	TargetModelName string           `yaml:"target_model_name" json:"target_model_name"`
	Type            RelationshipType `yaml:"type" json:"type"`
	ForeignKey      string           `yaml:"foreign_key,omitempty" json:"foreign_key,omitempty"` // Field in TargetModel pointing to SourceModel
	References      string           `yaml:"references,omitempty" json:"references,omitempty"`  // Field in SourceModel that ForeignKey references
	JoinTable       string           `yaml:"join_table,omitempty" json:"join_table,omitempty"`  // For many2many
	ThroughModel    string           `yaml:"through_model,omitempty" json:"through_model,omitempty"` // For many2many "through"
	FieldName       string           `yaml:"field_name,omitempty" json:"field_name,omitempty"` // Name of the field in SourceModel that defines this relationship
}
