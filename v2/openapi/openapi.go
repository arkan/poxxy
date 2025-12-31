package openapi

import (
	"encoding/json"

	"github.com/arkan/poxxy/v2"
)

// Schema represents an OpenAPI 3.0 Schema Object.
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Minimum     *float64           `json:"minimum,omitempty"`
	Maximum     *float64           `json:"maximum,omitempty"`
	MinLength   *int               `json:"minLength,omitempty"`
	MaxLength   *int               `json:"maxLength,omitempty"`
	MinItems    *int               `json:"minItems,omitempty"`
	MaxItems    *int               `json:"maxItems,omitempty"`
	Pattern     string             `json:"pattern,omitempty"`
	Enum        []any              `json:"enum,omitempty"`
	Default     any                `json:"default,omitempty"`
	Nullable    bool               `json:"nullable,omitempty"`
	AdditionalProperties *Schema   `json:"additionalProperties,omitempty"`
}

// Generator generates OpenAPI schemas from Poxxy schemas.
type Generator struct{}

// NewGenerator creates a new OpenAPI generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// FromSchema generates an OpenAPI Schema from a Poxxy Schema.
func (g *Generator) FromSchema(s *poxxy.Schema) *Schema {
	fieldInfos := s.FieldInfos()
	return g.fromFieldInfos(fieldInfos)
}

// fromFieldInfos generates an OpenAPI Schema from field infos.
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

// fromFieldInfo generates an OpenAPI Schema from a single field info.
func (g *Generator) fromFieldInfo(info poxxy.FieldInfo) *Schema {
	schema := &Schema{
		Description: info.Description,
		Default:     info.Default,
		Nullable:    info.Nullable,
	}

	// Set type based on field info
	if info.IsStruct {
		schema.Type = "object"
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
		schema.Type = "array"
		schema.Items = g.itemSchemaFromType(info.Type)
	} else if info.IsMap {
		schema.Type = "object"
		schema.AdditionalProperties = g.valueSchemaFromMapType(info.Type)
	} else {
		g.setTypeFromGoType(schema, info.Type)
	}

	// Apply validator constraints
	g.applyValidatorConstraints(schema, info.Validators)

	return schema
}

// setTypeFromGoType sets the OpenAPI type/format from a Go type name.
func (g *Generator) setTypeFromGoType(schema *Schema, goType string) {
	switch goType {
	case "string", "*string":
		schema.Type = "string"
	case "int", "int32", "*int", "*int32":
		schema.Type = "integer"
		schema.Format = "int32"
	case "int64", "*int64":
		schema.Type = "integer"
		schema.Format = "int64"
	case "float32", "*float32":
		schema.Type = "number"
		schema.Format = "float"
	case "float64", "*float64":
		schema.Type = "number"
		schema.Format = "double"
	case "bool", "*bool":
		schema.Type = "boolean"
	default:
		schema.Type = "string"
	}
}

// itemSchemaFromType extracts item schema from slice type.
func (g *Generator) itemSchemaFromType(sliceType string) *Schema {
	// Extract element type from []T
	if len(sliceType) > 2 && sliceType[:2] == "[]" {
		elemType := sliceType[2:]
		schema := &Schema{}
		g.setTypeFromGoType(schema, elemType)
		return schema
	}
	return &Schema{Type: "string"}
}

// valueSchemaFromMapType extracts value schema from map type.
func (g *Generator) valueSchemaFromMapType(mapType string) *Schema {
	// Simple extraction - assumes map[string]V format
	schema := &Schema{}
	schema.Type = "string" // Default
	return schema
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
