package poxxy

import (
	"strings"
	"testing"
)

func TestBasicFieldParsing(t *testing.T) {
	var name string
	var age int

	schema := NewSchema(
		Field("name", &name),
		Field("age", &age),
	)

	input := `{"name": "John", "age": 30}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "John" {
		t.Errorf("expected name='John', got %q", name)
	}
	if age != 30 {
		t.Errorf("expected age=30, got %d", age)
	}
}

func TestDefaultValue(t *testing.T) {
	var status string

	schema := NewSchema(
		Field("status", &status).Default("pending"),
	)

	input := `{}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != "pending" {
		t.Errorf("expected status='pending', got %q", status)
	}
}

func TestDefaultOnEmpty(t *testing.T) {
	var name string

	schema := NewSchema(
		Field("name", &name).DefaultOnEmpty("Anonymous"),
	)

	input := `{"name": ""}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "Anonymous" {
		t.Errorf("expected name='Anonymous', got %q", name)
	}
}

func TestPointerField(t *testing.T) {
	var nickname *string

	schema := NewSchema(
		Pointer("nickname", &nickname),
	)

	t.Run("with value", func(t *testing.T) {
		nickname = nil
		input := `{"nickname": "Johnny"}`
		err := schema.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if nickname == nil {
			t.Fatal("expected nickname to be set")
		}
		if *nickname != "Johnny" {
			t.Errorf("expected nickname='Johnny', got %q", *nickname)
		}
	})

	t.Run("absent", func(t *testing.T) {
		nickname = nil
		input := `{}`
		err := schema.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if nickname != nil {
			t.Error("expected nickname to be nil")
		}
	})
}

func TestSliceField(t *testing.T) {
	var tags []string

	schema := NewSchema(
		Slice("tags", &tags),
	)

	input := `{"tags": ["go", "validation", "schema"]}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"go", "validation", "schema"}
	if len(tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d", len(expected), len(tags))
	}
	for i, tag := range tags {
		if tag != expected[i] {
			t.Errorf("expected tags[%d]=%q, got %q", i, expected[i], tag)
		}
	}
}

func TestMapField(t *testing.T) {
	var metadata map[string]string

	schema := NewSchema(
		Map("metadata", &metadata),
	)

	input := `{"metadata": {"key1": "value1", "key2": "value2"}}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metadata["key1"] != "value1" {
		t.Errorf("expected metadata[key1]='value1', got %q", metadata["key1"])
	}
	if metadata["key2"] != "value2" {
		t.Errorf("expected metadata[key2]='value2', got %q", metadata["key2"])
	}
}

func TestNestedStruct(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}

	var address Address

	schema := NewSchema(
		Struct("address", &address, func(s *Schema) {
			s.Add(Field("street", &address.Street))
			s.Add(Field("city", &address.City))
		}),
	)

	input := `{"address": {"street": "123 Main St", "city": "Springfield"}}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if address.Street != "123 Main St" {
		t.Errorf("expected street='123 Main St', got %q", address.Street)
	}
	if address.City != "Springfield" {
		t.Errorf("expected city='Springfield', got %q", address.City)
	}
}

func TestTypeConversion(t *testing.T) {
	var count int

	schema := NewSchema(
		Field("count", &count),
	)

	// JSON numbers are float64, should convert to int
	input := `{"count": 42}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 42 {
		t.Errorf("expected count=42, got %d", count)
	}
}

func TestStrictTypeMode(t *testing.T) {
	var count int

	schema := NewSchema(
		Field("count", &count).StrictType(),
	)

	// String should fail in strict mode
	input := `{"count": "42"}`
	err := schema.Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for string to int in strict mode")
	}
}

func TestIsAbsentAndIsNull(t *testing.T) {
	var name string

	f := Field("name", &name)
	schema := NewSchema(f)

	t.Run("absent", func(t *testing.T) {
		input := `{}`
		err := schema.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !f.isAbsent() {
			t.Error("expected field to be absent")
		}
		if f.isNull() {
			t.Error("expected field not to be null")
		}
	})
}

func TestConvertField(t *testing.T) {
	var id int

	schema := NewSchema(
		Convert("id", &id, func(s string) (int, error) {
			// Custom conversion from string prefix
			if strings.HasPrefix(s, "ID-") {
				s = s[3:]
			}
			var result int
			_, err := strings.NewReader(s).Read([]byte{})
			if err != nil {
				return 0, err
			}
			// Simple conversion
			for _, c := range s {
				if c >= '0' && c <= '9' {
					result = result*10 + int(c-'0')
				}
			}
			return result, nil
		}),
	)

	input := `{"id": "ID-123"}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 123 {
		t.Errorf("expected id=123, got %d", id)
	}
}

func TestSliceWithCallback(t *testing.T) {
	type User struct {
		Name  string
		Email string
	}

	var users []User

	schema := NewSchema(
		Slice("users", &users, func(u *User, s *Schema) {
			s.Add(Field("name", &u.Name))
			s.Add(Field("email", &u.Email))
		}),
	)

	input := `{"users": [{"name": "Alice", "email": "alice@example.com"}, {"name": "Bob", "email": "bob@example.com"}]}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected users[0].Name='Alice', got %q", users[0].Name)
	}
	if users[1].Email != "bob@example.com" {
		t.Errorf("expected users[1].Email='bob@example.com', got %q", users[1].Email)
	}
}
