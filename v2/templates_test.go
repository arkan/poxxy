package poxxy

import (
	"strings"
	"testing"
)

func TestFieldTemplate(t *testing.T) {
	tmpl := Template[string]().
		Transform(TrimSpace(), ToLower()).
		Validate(Required[string](), MinLength[string](3))

	var email1, email2 string

	schema := NewSchema(
		tmpl.Bind("email", &email1),
		tmpl.Bind("backup_email", &email2),
	)

	err := schema.Parse(strings.NewReader(`{
		"email": "  TEST@EXAMPLE.COM  ",
		"backup_email": "  BACKUP@EXAMPLE.COM  "
	}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email1 != "test@example.com" {
		t.Errorf("email1: expected 'test@example.com', got %q", email1)
	}
	if email2 != "backup@example.com" {
		t.Errorf("email2: expected 'backup@example.com', got %q", email2)
	}
}

func TestFieldTemplateValidation(t *testing.T) {
	tmpl := Template[string]().
		Transform(TrimSpace()).
		Validate(Required[string](), MinLength[string](5))

	var value string

	schema := NewSchema(
		tmpl.Bind("value", &value),
	)

	err := schema.Parse(strings.NewReader(`{"value": "  ab  "}`))

	if err == nil {
		t.Fatal("expected validation error")
	}

	verrs, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	if len(verrs.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(verrs.Errors))
	}
}

func TestFieldTemplateDefault(t *testing.T) {
	tmpl := Template[string]().
		Default("default_value")

	var value string

	schema := NewSchema(
		tmpl.Bind("value", &value),
	)

	err := schema.Parse(strings.NewReader(`{}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != "default_value" {
		t.Errorf("expected 'default_value', got %q", value)
	}
}

func TestFieldTemplateDefaultOnEmpty(t *testing.T) {
	tmpl := Template[string]().
		DefaultOnEmpty("default_value")

	var value string

	schema := NewSchema(
		tmpl.Bind("value", &value),
	)

	err := schema.Parse(strings.NewReader(`{"value": ""}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != "default_value" {
		t.Errorf("expected 'default_value', got %q", value)
	}
}

func TestFieldTemplateStrictType(t *testing.T) {
	tmpl := Template[int]().
		StrictType()

	var value int

	schema := NewSchema(
		tmpl.Bind("value", &value),
	)

	// String should fail with strict type
	err := schema.Parse(strings.NewReader(`{"value": "123"}`))

	if err == nil {
		t.Fatal("expected type error with strict mode")
	}
}

func TestFieldTemplateDescribe(t *testing.T) {
	tmpl := Template[string]().
		Describe("A test field")

	var value string

	field := tmpl.Bind("value", &value)

	if field.desc != "A test field" {
		t.Errorf("expected description 'A test field', got %q", field.desc)
	}
}

func TestPointerTemplate(t *testing.T) {
	tmpl := PtrTemplate[string]().
		Transform(TrimSpace(), ToLower()).
		Validate(Email())

	var email *string

	schema := NewSchema(
		tmpl.Bind("email", &email),
	)

	err := schema.Parse(strings.NewReader(`{"email": "  TEST@EXAMPLE.COM  "}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email == nil {
		t.Fatal("expected email to be set")
	}
	if *email != "test@example.com" {
		t.Errorf("expected 'test@example.com', got %q", *email)
	}
}

func TestPointerTemplateNil(t *testing.T) {
	tmpl := PtrTemplate[string]().
		Validate(Email())

	var email *string

	schema := NewSchema(
		tmpl.Bind("email", &email),
	)

	err := schema.Parse(strings.NewReader(`{}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email != nil {
		t.Errorf("expected email to be nil, got %v", email)
	}
}

func TestSliceTemplate(t *testing.T) {
	tmpl := SliceTempl[string]().
		Validate(MinItems[string](1), MaxItems[string](5)).
		Describe("List of tags")

	var tags []string

	schema := NewSchema(
		tmpl.Bind("tags", &tags),
	)

	err := schema.Parse(strings.NewReader(`{"tags": ["go", "validation"]}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestSliceTemplateWithItemSchema(t *testing.T) {
	type User struct {
		Name  string
		Email string
	}

	tmpl := SliceTempl[User]().
		Validate(MinItems[User](1)).
		WithItemSchema(func(u *User, s *Schema) {
			s.Add(Field("name", &u.Name).Validate(Required[string]()))
			s.Add(Field("email", &u.Email).Validate(Required[string](), Email()))
		})

	var users []User

	schema := NewSchema(
		tmpl.Bind("users", &users),
	)

	err := schema.Parse(strings.NewReader(`{
		"users": [
			{"name": "John", "email": "john@example.com"},
			{"name": "Jane", "email": "jane@example.com"}
		]
	}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "John" {
		t.Errorf("expected first user name 'John', got %q", users[0].Name)
	}
}

func TestStructTemplate(t *testing.T) {
	type Address struct {
		Street string
		City   string
		Zip    string
	}

	tmpl := StructTempl[Address](func(a *Address, s *Schema) {
		s.Add(Field("street", &a.Street).Validate(Required[string]()))
		s.Add(Field("city", &a.City).Validate(Required[string]()))
		s.Add(Field("zip", &a.Zip).Validate(Required[string](), Pattern(`^\d{5}$`)))
	}).Describe("Address information")

	var addr Address

	schema := NewSchema(
		tmpl.Bind("address", &addr),
	)

	err := schema.Parse(strings.NewReader(`{
		"address": {
			"street": "123 Main St",
			"city": "Springfield",
			"zip": "12345"
		}
	}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addr.Street != "123 Main St" {
		t.Errorf("expected street '123 Main St', got %q", addr.Street)
	}
	if addr.City != "Springfield" {
		t.Errorf("expected city 'Springfield', got %q", addr.City)
	}
	if addr.Zip != "12345" {
		t.Errorf("expected zip '12345', got %q", addr.Zip)
	}
}

func TestStructTemplateReuse(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}

	tmpl := StructTempl[Address](func(a *Address, s *Schema) {
		s.Add(Field("street", &a.Street).Validate(Required[string]()))
		s.Add(Field("city", &a.City).Validate(Required[string]()))
	})

	var homeAddr, workAddr Address

	schema := NewSchema(
		tmpl.Bind("home_address", &homeAddr),
		tmpl.Bind("work_address", &workAddr),
	)

	err := schema.Parse(strings.NewReader(`{
		"home_address": {"street": "123 Home St", "city": "HomeTown"},
		"work_address": {"street": "456 Work Ave", "city": "WorkCity"}
	}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if homeAddr.Street != "123 Home St" {
		t.Errorf("expected home street '123 Home St', got %q", homeAddr.Street)
	}
	if workAddr.Street != "456 Work Ave" {
		t.Errorf("expected work street '456 Work Ave', got %q", workAddr.Street)
	}
}

func TestEmailTemplatePreset(t *testing.T) {
	var email string

	schema := NewSchema(
		EmailTemplate.Bind("email", &email),
	)

	err := schema.Parse(strings.NewReader(`{"email": "  TEST@EXAMPLE.COM  "}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email != "test@example.com" {
		t.Errorf("expected 'test@example.com', got %q", email)
	}
}

func TestRequiredEmailTemplatePreset(t *testing.T) {
	var email string

	schema := NewSchema(
		RequiredEmailTemplate.Bind("email", &email),
	)

	// Missing required email
	err := schema.Parse(strings.NewReader(`{}`))

	if err == nil {
		t.Fatal("expected validation error for missing required email")
	}
}

func TestURLTemplatePreset(t *testing.T) {
	var url string

	schema := NewSchema(
		URLTemplate.Bind("website", &url),
	)

	err := schema.Parse(strings.NewReader(`{"website": "  https://example.com  "}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if url != "https://example.com" {
		t.Errorf("expected 'https://example.com', got %q", url)
	}
}

func TestUUIDTemplatePreset(t *testing.T) {
	var id string

	schema := NewSchema(
		UUIDTemplate.Bind("id", &id),
	)

	err := schema.Parse(strings.NewReader(`{"id": "  550E8400-E29B-41D4-A716-446655440000  "}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected lowercase uuid, got %q", id)
	}
}

func TestPositiveIntTemplatePreset(t *testing.T) {
	var count int

	schema := NewSchema(
		PositiveIntTemplate.Bind("count", &count),
	)

	// Zero should fail
	err := schema.Parse(strings.NewReader(`{"count": 0}`))

	if err == nil {
		t.Fatal("expected validation error for zero value with PositiveIntTemplate")
	}

	// Positive should pass
	err = schema.Parse(strings.NewReader(`{"count": 5}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

func TestNonNegativeIntTemplatePreset(t *testing.T) {
	var count int

	schema := NewSchema(
		NonNegativeIntTemplate.Bind("count", &count),
	)

	// Zero should pass
	err := schema.Parse(strings.NewReader(`{"count": 0}`))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// Negative should fail
	err = schema.Parse(strings.NewReader(`{"count": -1}`))

	if err == nil {
		t.Fatal("expected validation error for negative value with NonNegativeIntTemplate")
	}
}

func TestTemplateImmutability(t *testing.T) {
	// Templates should be immutable - binding shouldn't affect the original template
	tmpl := Template[string]().
		Validate(Required[string]())

	var value1, value2 string

	// Create two fields from the same template
	field1 := tmpl.Bind("value1", &value1)
	field2 := tmpl.Bind("value2", &value2)

	// Modify field1 after binding
	field1.Validate(MinLength[string](5))

	// field2 should not be affected
	if len(field2.validators) != len(field1.validators)-1 {
		// Note: field1 has one extra validator (MinLength) that was added after binding
		// field2 should only have the original Required validator from the template
	}

	// Both should work independently
	schema1 := NewSchema(field1)
	schema2 := NewSchema(field2)

	// field1 needs at least 5 chars
	err := schema1.Parse(strings.NewReader(`{"value1": "ab"}`))
	if err == nil {
		t.Error("field1 should require MinLength(5)")
	}

	// field2 just needs to be non-empty
	err = schema2.Parse(strings.NewReader(`{"value2": "ab"}`))
	if err != nil {
		t.Errorf("field2 should pass with 'ab': %v", err)
	}
}
