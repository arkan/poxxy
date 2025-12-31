package openapi

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
	openAPI := gen.FromSchema(schema)

	if openAPI.Type != "object" {
		t.Errorf("expected type 'object', got %q", openAPI.Type)
	}

	if len(openAPI.Properties) != 4 {
		t.Errorf("expected 4 properties, got %d", len(openAPI.Properties))
	}

	// Check name property
	nameProp := openAPI.Properties["name"]
	if nameProp == nil {
		t.Fatal("expected 'name' property")
	}
	if nameProp.Type != "string" {
		t.Errorf("expected name type 'string', got %q", nameProp.Type)
	}
	if nameProp.Description != "User name" {
		t.Errorf("expected description 'User name', got %q", nameProp.Description)
	}

	// Check age property
	ageProp := openAPI.Properties["age"]
	if ageProp == nil {
		t.Fatal("expected 'age' property")
	}
	if ageProp.Type != "integer" {
		t.Errorf("expected age type 'integer', got %q", ageProp.Type)
	}

	// Check score property
	scoreProp := openAPI.Properties["score"]
	if scoreProp == nil {
		t.Fatal("expected 'score' property")
	}
	if scoreProp.Type != "number" {
		t.Errorf("expected score type 'number', got %q", scoreProp.Type)
	}

	// Check admin property
	adminProp := openAPI.Properties["admin"]
	if adminProp == nil {
		t.Fatal("expected 'admin' property")
	}
	if adminProp.Type != "boolean" {
		t.Errorf("expected admin type 'boolean', got %q", adminProp.Type)
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
	openAPI := gen.FromSchema(schema)

	if len(openAPI.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(openAPI.Required))
	}

	// Check email format
	emailProp := openAPI.Properties["email"]
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
	openAPI := gen.FromSchema(schema)

	// Check name constraints
	nameProp := openAPI.Properties["name"]
	if nameProp.MinLength == nil || *nameProp.MinLength != 3 {
		t.Error("expected minLength 3 for name")
	}
	if nameProp.MaxLength == nil || *nameProp.MaxLength != 50 {
		t.Error("expected maxLength 50 for name")
	}

	// Check age constraints
	ageProp := openAPI.Properties["age"]
	if ageProp.Minimum == nil || *ageProp.Minimum != 18 {
		t.Error("expected minimum 18 for age")
	}
	if ageProp.Maximum == nil || *ageProp.Maximum != 120 {
		t.Error("expected maximum 120 for age")
	}
}

func TestGenerator_SliceField(t *testing.T) {
	var tags []string

	schema := poxxy.NewSchema(
		poxxy.Slice("tags", &tags).Validate(poxxy.MinItems[string](1), poxxy.MaxItems[string](10)),
	)

	gen := NewGenerator()
	openAPI := gen.FromSchema(schema)

	tagsProp := openAPI.Properties["tags"]
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

func TestGenerator_MapField(t *testing.T) {
	var metadata map[string]string

	schema := poxxy.NewSchema(
		poxxy.Map("metadata", &metadata),
	)

	gen := NewGenerator()
	openAPI := gen.FromSchema(schema)

	metaProp := openAPI.Properties["metadata"]
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
	openAPI := gen.FromSchema(schema)

	nickProp := openAPI.Properties["nickname"]
	if nickProp == nil {
		t.Fatal("expected 'nickname' property")
	}
	if !nickProp.Nullable {
		t.Error("expected nullable true for pointer field")
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
	openAPI := gen.FromSchema(schema)

	addrProp := openAPI.Properties["address"]
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
	openAPI := gen.FromSchema(schema)

	statusProp := openAPI.Properties["status"]
	if statusProp == nil {
		t.Fatal("expected 'status' property")
	}
	if len(statusProp.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(statusProp.Enum))
	}
}

func TestGenerator_FormatConstraints(t *testing.T) {
	var (
		email   string
		website string
		id      string
	)

	schema := poxxy.NewSchema(
		poxxy.Field("email", &email).Validate(poxxy.Email()),
		poxxy.Field("website", &website).Validate(poxxy.URL()),
		poxxy.Field("id", &id).Validate(poxxy.UUID()),
	)

	gen := NewGenerator()
	openAPI := gen.FromSchema(schema)

	if openAPI.Properties["email"].Format != "email" {
		t.Errorf("expected format 'email', got %q", openAPI.Properties["email"].Format)
	}
	if openAPI.Properties["website"].Format != "uri" {
		t.Errorf("expected format 'uri', got %q", openAPI.Properties["website"].Format)
	}
	if openAPI.Properties["id"].Format != "uuid" {
		t.Errorf("expected format 'uuid', got %q", openAPI.Properties["id"].Format)
	}
}

func TestGenerator_DefaultValue(t *testing.T) {
	var role string

	schema := poxxy.NewSchema(
		poxxy.Field("role", &role).Default("user"),
	)

	gen := NewGenerator()
	openAPI := gen.FromSchema(schema)

	roleProp := openAPI.Properties["role"]
	if roleProp.Default != "user" {
		t.Errorf("expected default 'user', got %v", roleProp.Default)
	}
}

func TestSchema_JSON(t *testing.T) {
	var name string

	schema := poxxy.NewSchema(
		poxxy.Field("name", &name).Validate(poxxy.Required[string]()),
	)

	gen := NewGenerator()
	openAPI := gen.FromSchema(schema)

	data, err := openAPI.JSON()
	if err != nil {
		t.Fatalf("failed to generate JSON: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result["type"] != "object" {
		t.Errorf("expected type 'object', got %v", result["type"])
	}
}
