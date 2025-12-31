package jsonschema

import (
	"encoding/json"

	"github.com/arkan/poxxy/v2"
)

// Schema represents a JSON Schema (Draft-07).
type Schema struct {
	Schema               string             `json:"$schema,omitempty"`
	ID                   string             `json:"$id,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Type                 any                `json:"type,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
	ExclusiveMinimum     *float64           `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *float64           `json:"exclusiveMaximum,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	MinItems             *int               `json:"minItems,omitempty"`
	MaxItems             *int               `json:"maxItems,omitempty"`
	UniqueItems          bool               `json:"uniqueItems,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	Format               string             `json:"format,omitempty"`
	Enum                 []any              `json:"enum,omitempty"`
	Default              any                `json:"default,omitempty"`
	Const                any                `json:"const,omitempty"`
}

// Generator generates JSON Schema from Poxxy schemas.
type Generator struct {
	schemaURL string
	id        string
	title     string
}

// Option configures the generator.
type Option func(*Generator)

// WithSchemaURL sets the $schema URL (default: Draft-07).
func WithSchemaURL(url string) Option {
	return func(g *Generator) {
		g.schemaURL = url
	}
}

// WithID sets the $id of the generated schema.
func WithID(id string) Option {
	return func(g *Generator) {
		g.id = id
	}
}

// WithTitle sets the title of the generated schema.
func WithTitle(title string) Option {
	return func(g *Generator) {
		g.title = title
	}
}

// NewGenerator creates a new JSON Schema generator.
func NewGenerator(opts ...Option) *Generator {
	g := &Generator{
		schemaURL: "http://json-schema.org/draft-07/schema#",
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// FromSchema generates a JSON Schema from a Poxxy Schema.
func (g *Generator) FromSchema(s *poxxy.Schema) *Schema {
	fieldInfos := s.FieldInfos()
	schema := g.fromFieldInfos(fieldInfos)
	schema.Schema = g.schemaURL
	schema.ID = g.id
	schema.Title = g.title
	return schema
}

// fromFieldInfos generates a JSON Schema from field infos.
func (g *Generator) fromFieldInfos(infos []poxxy.FieldInfo) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   []string{},
	}

	for _, info := range infos {
		propSchema := g.fromFieldInfo(info)
		schema.Properties[info.Name] = propSchema

		if info.Required {
			schema.Required = append(schema.Required, info.Name)
		}
	}

	if len(schema.Required) == 0 {
		schema.Required = nil
	}

	return schema
}

// fromFieldInfo generates a JSON Schema from a single field info.
func (g *Generator) fromFieldInfo(info poxxy.FieldInfo) *Schema {
	schema := &Schema{
		Description: info.Description,
		Default:     info.Default,
	}

	// Handle nullable types
	if info.Nullable {
		schema.Type = []string{g.goTypeToJSONType(info.Type), "null"}
	}

	// Set type based on field info
	if info.IsStruct {
		if !info.Nullable {
			schema.Type = "object"
		}
		schema.Properties = make(map[string]*Schema)
		for _, child := range info.Children {
			schema.Properties[child.Name] = g.fromFieldInfo(child)
			if child.Required {
				schema.Required = append(schema.Required, child.Name)
			}
		}
		if len(schema.Required) == 0 {
			schema.Required = nil
		}
	} else if info.IsSlice {
		if !info.Nullable {
			schema.Type = "array"
		}
		schema.Items = g.itemSchemaFromType(info.Type)
	} else if info.IsMap {
		if !info.Nullable {
			schema.Type = "object"
		}
		schema.AdditionalProperties = g.valueSchemaFromMapType(info.Type)
	} else if !info.Nullable {
		schema.Type = g.goTypeToJSONType(info.Type)
	}

	// Apply validator constraints
	g.applyValidatorConstraints(schema, info.Validators)

	return schema
}

// goTypeToJSONType converts a Go type name to JSON Schema type.
func (g *Generator) goTypeToJSONType(goType string) string {
	switch goType {
	case "string", "*string":
		return "string"
	case "int", "int8", "int16", "int32", "int64",
		"*int", "*int8", "*int16", "*int32", "*int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"*uint", "*uint8", "*uint16", "*uint32", "*uint64":
		return "integer"
	case "float32", "float64", "*float32", "*float64":
		return "number"
	case "bool", "*bool":
		return "boolean"
	default:
		return "string"
	}
}

// itemSchemaFromType extracts item schema from slice type.
func (g *Generator) itemSchemaFromType(sliceType string) *Schema {
	// Extract element type from []T
	if len(sliceType) > 2 && sliceType[:2] == "[]" {
		elemType := sliceType[2:]
		return &Schema{Type: g.goTypeToJSONType(elemType)}
	}
	return &Schema{Type: "string"}
}

// valueSchemaFromMapType extracts value schema from map type.
func (g *Generator) valueSchemaFromMapType(mapType string) *Schema {
	return &Schema{Type: "string"} // Default
}

// applyValidatorConstraints applies validator constraints to the schema.
func (g *Generator) applyValidatorConstraints(schema *Schema, validators []poxxy.ValidatorInfo) {
	for _, v := range validators {
		switch v.Rule {
		case "min":
			if val, ok := getFloat(v.Params["min"]); ok {
				schema.Minimum = &val
			}
		case "max":
			if val, ok := getFloat(v.Params["max"]); ok {
				schema.Maximum = &val
			}
		case "positive":
			val := float64(0)
			schema.ExclusiveMinimum = &val
		case "negative":
			val := float64(0)
			schema.ExclusiveMaximum = &val
		case "min_length":
			if val, ok := getInt(v.Params["min"]); ok {
				schema.MinLength = &val
			}
		case "max_length":
			if val, ok := getInt(v.Params["max"]); ok {
				schema.MaxLength = &val
			}
		case "min_items":
			if val, ok := getInt(v.Params["min"]); ok {
				schema.MinItems = &val
			}
		case "max_items":
			if val, ok := getInt(v.Params["max"]); ok {
				schema.MaxItems = &val
			}
		case "pattern":
			if pattern, ok := v.Params["pattern"].(string); ok {
				schema.Pattern = pattern
			}
		case "in":
			if values, ok := v.Params["allowed"]; ok {
				schema.Enum = toAnySlice(values)
			}
		case "unique":
			schema.UniqueItems = true
		case "email":
			schema.Format = "email"
		case "url":
			schema.Format = "uri"
		case "uuid":
			schema.Format = "uuid"
		}
	}
}

// JSON returns the schema as JSON bytes.
func (s *Schema) JSON() ([]byte, error) {
	return json.Marshal(s)
}

// JSONIndent returns the schema as indented JSON bytes.
func (s *Schema) JSONIndent(prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(s, prefix, indent)
}

// Helper functions

func getFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	}
	return 0, false
}

func getInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case int32:
		return int(val), true
	case float64:
		return int(val), true
	}
	return 0, false
}

func toAnySlice(v any) []any {
	switch val := v.(type) {
	case []any:
		return val
	case []string:
		result := make([]any, len(val))
		for i, s := range val {
			result[i] = s
		}
		return result
	case []int:
		result := make([]any, len(val))
		for i, n := range val {
			result[i] = n
		}
		return result
	}
	return nil
}
