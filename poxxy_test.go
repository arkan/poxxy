package poxxy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestPoxxy_BasicTypes(t *testing.T) {
	var name string
	var age int
	var isAdmin bool
	var tags []string

	schema := NewSchema(
		Value("name", &name, WithValidators(Required())),
		Value("age", &age, WithValidators(Required())),
		Value("isAdmin", &isAdmin, WithValidators(Required())),
		Slice("tags", &tags, WithValidators(Required())),
	)

	jsonData := `{"name": "test", "age": 20, "isAdmin": true, "tags": ["tag1", "tag2"]}`
	err := schema.ApplyJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Schema.ApplyJSON() error = %v", err)
	}

	if name != "test" {
		t.Errorf("Schema.ApplyJSON() name = %v, want %v", name, "test")
	}
	if age != 20 {
		t.Errorf("Schema.ApplyJSON() age = %v, want %v", age, 20)
	}
	if !isAdmin {
		t.Errorf("Schema.ApplyJSON() isAdmin = %v, want %v", isAdmin, true)
	}
	if len(tags) != 2 || !reflect.DeepEqual(tags, []string{"tag1", "tag2"}) {
		t.Errorf("Schema.ApplyJSON() tags = %v, want %v", tags, []string{"tag1", "tag2"})
	}
}

func TestPoxxy_Map(t *testing.T) {
	type User struct {
		Name        string
		Age         int
		Preferences map[string]string
	}

	var user User
	schema := NewSchema(
		Value("name", &user.Name, WithValidators(Required())),
		Value("age", &user.Age, WithValidators(Required())),
		Map("preferences",
			&user.Preferences,
			WithValidators(Required()),
			WithMapCallback(func(schema *Schema, key string, value string) {
				WithSchema(schema, ValueWithoutAssign[string]("color", WithValidators(Required(), ValidatorFunc(func(value interface{}, fieldName string) error {
					color, ok := value.(string)
					if !ok {
						return fmt.Errorf("invalid color")
					}

					if !strings.HasPrefix(color, "b") {
						return fmt.Errorf("color must start with b for field %s", fieldName)
					}

					return nil
				}))))
				WithSchema(schema, ValueWithoutAssign[string]("size", WithValidators(Required(), In("small", "medium", "large"))))
			}),
		),
	)

	jsonData := `{"name": "test", "age": 20, "preferences": {"color": "blue", "size": "large"}}`
	err := schema.ApplyJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Schema.ApplyJSON() error = %v", err)
	}

	if user.Name != "test" {
		t.Errorf("Schema.ApplyJSON() name = %v, want %v", user.Name, "test")
	}
	if user.Age != 20 {
		t.Errorf("Schema.ApplyJSON() age = %v, want %v", user.Age, 20)
	}
	expectedPreferences := map[string]string{"color": "blue", "size": "large"}
	if !reflect.DeepEqual(user.Preferences, expectedPreferences) {
		t.Errorf("Schema.ApplyJSON() preferences = %v, want %v", user.Preferences, expectedPreferences)
	}
}

func TestPoxxy_Slice(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	var users []User
	schema := NewSchema(
		SliceOf("users", &users, func(schema *Schema, user *User) {
			WithSchema(schema, Value("name", &user.Name, WithValidators(Required())))
			WithSchema(schema, Value("age", &user.Age, WithValidators(Required())))
		}, WithValidators(Required())),
	)

	jsonData := `{"users": [{"name": "test", "age": 20}, {"name": "test2", "age": 21}]}`
	err := schema.ApplyJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Schema.ApplyJSON() error = %v", err)
	}

	expectedUsers := []User{{Name: "test", Age: 20}, {Name: "test2", Age: 21}}
	if len(users) != 2 {
		t.Errorf("Schema.ApplyJSON() users = %v, want %v", users, expectedUsers)
	}

	if users[0].Name != "test" {
		t.Errorf("Schema.ApplyJSON() users[0].Name = %v, want %v", users[0].Name, "test")
	}
	if users[0].Age != 20 {
		t.Errorf("Schema.ApplyJSON() users[0].Age = %v, want %v", users[0].Age, 20)
	}

	if users[1].Name != "test2" {
		t.Errorf("Schema.ApplyJSON() users[1].Name = %v, want %v", users[1].Name, "test2")
	}
	if users[1].Age != 21 {
		t.Errorf("Schema.ApplyJSON() users[1].Age = %v, want %v", users[1].Age, 21)
	}
}
