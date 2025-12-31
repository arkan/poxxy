package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/arkan/poxxy/v2"
)

func TestGenerator_BasicTypes(t *testing.T) {
	var (
		name  string
		age   int
		score float64
		admin bool
	)

	schema := poxxy.NewSchema(
		poxxy.Field("name", &name).Describe("User name"),
		poxxy.Field("age", &age).Describe("User age"),
		poxxy.Field("score", &score),
		poxxy.Field("admin", &admin),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	if jsonSchema.Type != "object" {
		t.Errorf("expected type 'object', got %q", jsonSchema.Type)
	}

	if jsonSchema.Schema != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("unexpected schema URL: %q", jsonSchema.Schema)
	}

	if len(jsonSchema.Properties) != 4 {
		t.Errorf("expected 4 properties, got %d", len(jsonSchema.Properties))
	}

	// Check name property
	nameProp := jsonSchema.Properties["name"]
	if nameProp == nil {
		t.Fatal("expected 'name' property")
	}
	if nameProp.Type != "string" {
		t.Errorf("expected name type 'string', got %q", nameProp.Type)
	}

	// Check age property
	ageProp := jsonSchema.Properties["age"]
	if ageProp == nil {
		t.Fatal("expected 'age' property")
	}
	if ageProp.Type != "integer" {
		t.Errorf("expected age type 'integer', got %q", ageProp.Type)
	}

	// Check score property
	scoreProp := jsonSchema.Properties["score"]
	if scoreProp == nil {
		t.Fatal("expected 'score' property")
	}
	if scoreProp.Type != "number" {
		t.Errorf("expected score type 'number', got %q", scoreProp.Type)
	}

	// Check admin property
	adminProp := jsonSchema.Properties["admin"]
	if adminProp == nil {
		t.Fatal("expected 'admin' property")
	}
	if adminProp.Type != "boolean" {
		t.Errorf("expected admin type 'boolean', got %q", adminProp.Type)
	}
}

func TestGenerator_WithOptions(t *testing.T) {
	var name string

	schema := poxxy.NewSchema(
		poxxy.Field("name", &name),
	)

	gen := NewGenerator(
		WithID("https://example.com/user.schema.json"),
		WithTitle("User Schema"),
	)
	jsonSchema := gen.FromSchema(schema)

	if jsonSchema.ID != "https://example.com/user.schema.json" {
		t.Errorf("unexpected $id: %q", jsonSchema.ID)
	}
	if jsonSchema.Title != "User Schema" {
		t.Errorf("unexpected title: %q", jsonSchema.Title)
	}
}

func TestGenerator_RequiredFields(t *testing.T) {
	var (
		name  string
		email string
	)

	schema := poxxy.NewSchema(
		poxxy.Field("name", &name).Validate(poxxy.Required[string]()),
		poxxy.Field("email", &email).Validate(poxxy.Required[string](), poxxy.Email()),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	if len(jsonSchema.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(jsonSchema.Required))
	}

	// Check email format
	emailProp := jsonSchema.Properties["email"]
	if emailProp.Format != "email" {
		t.Errorf("expected email format 'email', got %q", emailProp.Format)
	}
}

func TestGenerator_ValidatorConstraints(t *testing.T) {
	var (
		name string
		age  int
	)

	schema := poxxy.NewSchema(
		poxxy.Field("name", &name).Validate(poxxy.MinLength[string](3), poxxy.MaxLength[string](50)),
		poxxy.Field("age", &age).Validate(poxxy.Min(18), poxxy.Max(120)),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	// Check name constraints
	nameProp := jsonSchema.Properties["name"]
	if nameProp.MinLength == nil || *nameProp.MinLength != 3 {
		t.Error("expected minLength 3 for name")
	}
	if nameProp.MaxLength == nil || *nameProp.MaxLength != 50 {
		t.Error("expected maxLength 50 for name")
	}

	// Check age constraints
	ageProp := jsonSchema.Properties["age"]
	if ageProp.Minimum == nil || *ageProp.Minimum != 18 {
		t.Error("expected minimum 18 for age")
	}
	if ageProp.Maximum == nil || *ageProp.Maximum != 120 {
		t.Error("expected maximum 120 for age")
	}
}

func TestGenerator_PositiveNegativeConstraints(t *testing.T) {
	var (
		positiveNum int
		negativeNum int
	)

	schema := poxxy.NewSchema(
		poxxy.Field("positive", &positiveNum).Validate(poxxy.Positive[int]()),
		poxxy.Field("negative", &negativeNum).Validate(poxxy.Negative[int]()),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	// Check positive constraint
	posProp := jsonSchema.Properties["positive"]
	if posProp.ExclusiveMinimum == nil || *posProp.ExclusiveMinimum != 0 {
		t.Error("expected exclusiveMinimum 0 for positive")
	}

	// Check negative constraint
	negProp := jsonSchema.Properties["negative"]
	if negProp.ExclusiveMaximum == nil || *negProp.ExclusiveMaximum != 0 {
		t.Error("expected exclusiveMaximum 0 for negative")
	}
}

func TestGenerator_SliceField(t *testing.T) {
	var tags []string

	schema := poxxy.NewSchema(
		poxxy.Slice("tags", &tags).Validate(poxxy.MinItems[string](1), poxxy.MaxItems[string](10)),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	tagsProp := jsonSchema.Properties["tags"]
	if tagsProp == nil {
		t.Fatal("expected 'tags' property")
	}
	if tagsProp.Type != "array" {
		t.Errorf("expected type 'array', got %q", tagsProp.Type)
	}
	if tagsProp.Items == nil {
		t.Fatal("expected items schema")
	}
	if tagsProp.Items.Type != "string" {
		t.Errorf("expected items type 'string', got %q", tagsProp.Items.Type)
	}
	if tagsProp.MinItems == nil || *tagsProp.MinItems != 1 {
		t.Error("expected minItems 1")
	}
	if tagsProp.MaxItems == nil || *tagsProp.MaxItems != 10 {
		t.Error("expected maxItems 10")
	}
}

func TestGenerator_UniqueConstraint(t *testing.T) {
	var ids []int

	schema := poxxy.NewSchema(
		poxxy.Slice("ids", &ids).Validate(poxxy.Unique[int]()),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	idsProp := jsonSchema.Properties["ids"]
	if !idsProp.UniqueItems {
		t.Error("expected uniqueItems true")
	}
}

func TestGenerator_MapField(t *testing.T) {
	var metadata map[string]string

	schema := poxxy.NewSchema(
		poxxy.Map("metadata", &metadata),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	metaProp := jsonSchema.Properties["metadata"]
	if metaProp == nil {
		t.Fatal("expected 'metadata' property")
	}
	if metaProp.Type != "object" {
		t.Errorf("expected type 'object', got %q", metaProp.Type)
	}
}

func TestGenerator_PointerField(t *testing.T) {
	var nickname *string

	schema := poxxy.NewSchema(
		poxxy.Pointer("nickname", &nickname),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	nickProp := jsonSchema.Properties["nickname"]
	if nickProp == nil {
		t.Fatal("expected 'nickname' property")
	}

	// Check nullable type
	types, ok := nickProp.Type.([]string)
	if !ok {
		t.Fatal("expected type to be array for nullable field")
	}
	if len(types) != 2 {
		t.Errorf("expected 2 types, got %d", len(types))
	}
	hasString := false
	hasNull := false
	for _, tp := range types {
		if tp == "string" {
			hasString = true
		}
		if tp == "null" {
			hasNull = true
		}
	}
	if !hasString || !hasNull {
		t.Error("expected types to include 'string' and 'null'")
	}
}

func TestGenerator_NestedStruct(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}

	var (
		name    string
		address Address
	)

	schema := poxxy.NewSchema(
		poxxy.Field("name", &name).Validate(poxxy.Required[string]()),
		poxxy.Struct("address", &address, func(s *poxxy.Schema) {
			s.Add(poxxy.Field("street", &address.Street).Validate(poxxy.Required[string]()))
			s.Add(poxxy.Field("city", &address.City).Validate(poxxy.Required[string]()))
		}),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	addrProp := jsonSchema.Properties["address"]
	if addrProp == nil {
		t.Fatal("expected 'address' property")
	}
	if addrProp.Type != "object" {
		t.Errorf("expected type 'object', got %q", addrProp.Type)
	}
	if len(addrProp.Properties) != 2 {
		t.Errorf("expected 2 nested properties, got %d", len(addrProp.Properties))
	}
	if len(addrProp.Required) != 2 {
		t.Errorf("expected 2 required nested fields, got %d", len(addrProp.Required))
	}
}

func TestGenerator_EnumConstraint(t *testing.T) {
	var status string

	schema := poxxy.NewSchema(
		poxxy.Field("status", &status).Validate(poxxy.In("active", "inactive", "pending")),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	statusProp := jsonSchema.Properties["status"]
	if statusProp == nil {
		t.Fatal("expected 'status' property")
	}
	if len(statusProp.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(statusProp.Enum))
	}
}

func TestGenerator_PatternConstraint(t *testing.T) {
	var phone string

	schema := poxxy.NewSchema(
		poxxy.Field("phone", &phone).Validate(poxxy.Pattern(`^\+?[0-9]{10,14}$`)),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	phoneProp := jsonSchema.Properties["phone"]
	if phoneProp.Pattern != `^\+?[0-9]{10,14}$` {
		t.Errorf("unexpected pattern: %q", phoneProp.Pattern)
	}
}

func TestGenerator_DefaultValue(t *testing.T) {
	var role string

	schema := poxxy.NewSchema(
		poxxy.Field("role", &role).Default("user"),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	roleProp := jsonSchema.Properties["role"]
	if roleProp.Default != "user" {
		t.Errorf("expected default 'user', got %v", roleProp.Default)
	}
}

func TestSchema_JSON(t *testing.T) {
	var name string

	schema := poxxy.NewSchema(
		poxxy.Field("name", &name).Validate(poxxy.Required[string]()),
	)

	gen := NewGenerator(
		WithID("https://example.com/test.schema.json"),
		WithTitle("Test"),
	)
	jsonSchema := gen.FromSchema(schema)

	data, err := jsonSchema.JSON()
	if err != nil {
		t.Fatalf("failed to generate JSON: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result["$schema"] != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("unexpected $schema: %v", result["$schema"])
	}
	if result["$id"] != "https://example.com/test.schema.json" {
		t.Errorf("unexpected $id: %v", result["$id"])
	}
	if result["title"] != "Test" {
		t.Errorf("unexpected title: %v", result["title"])
	}
	if result["type"] != "object" {
		t.Errorf("expected type 'object', got %v", result["type"])
	}
}

func TestSchema_JSONIndent(t *testing.T) {
	var name string

	schema := poxxy.NewSchema(
		poxxy.Field("name", &name),
	)

	gen := NewGenerator()
	jsonSchema := gen.FromSchema(schema)

	data, err := jsonSchema.JSONIndent("", "  ")
	if err != nil {
		t.Fatalf("failed to generate JSON: %v", err)
	}

	// Should have indentation
	if len(data) < 50 {
		t.Error("expected indented JSON to be longer")
	}
}
