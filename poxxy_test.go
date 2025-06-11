package poxxy

import (
	"database/sql"
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
			WithSubSchemaMap(func(schema *Schema, key string, value string) {
				WithSchema(
					schema,
					ValueWithoutAssign[string]("color",
						WithValidators(
							Required(),
							ValidatorFunc(func(color string, fieldName string) error {
								if !strings.HasPrefix(color, "b") {
									return fmt.Errorf("color must start with b for field %s", fieldName)
								}

								return nil
							}),
						),
					),
				)
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
		Slice("users", &users,
			WithSubSchema(func(schema *Schema, user *User) {
				WithSchema(schema, Value("name", &user.Name, WithValidators(Required())))
				WithSchema(schema, Value("age", &user.Age, WithValidators(Required())))
			}),
			WithValidators(Required()),
		),
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

func TestPoxxy_Boolean(t *testing.T) {
	{
		var isAdmin bool
		schema := NewSchema(
			Value("isAdmin", &isAdmin, WithValidators(Required())),
		)

		jsonData := `{"isAdmin": true}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if !isAdmin {
			t.Errorf("Schema.ApplyJSON() isAdmin = %v, want %v", isAdmin, true)
		}
	}

	{
		var isAdmin *bool
		schema := NewSchema(
			Pointer("isAdmin", &isAdmin, WithValidators(Required())),
		)

		jsonData := `{"isAdmin": true}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if isAdmin != nil && *isAdmin != true {
			t.Errorf("Schema.ApplyJSON() isAdmin = %v, want %v", *isAdmin, true)
		}
	}

	{
		var isAdmin *bool
		schema := NewSchema(
			Pointer("isAdmin", &isAdmin, WithValidators(Required())),
		)

		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err == nil {
			t.Errorf("Schema.ApplyJSON() error = nil, want not nil")
		}

		if !strings.Contains(err.Error(), "isAdmin: field is required") {
			t.Errorf("Schema.ApplyJSON() error = %v, want %v", err, "isAdmin: field is required")
		}
	}
}

func TestPoxxy_Struct_Required(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	{
		var user User
		schema := NewSchema(
			Struct("user", &user, WithValidators(Required())),
		)

		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err == nil {
			t.Errorf("Schema.ApplyJSON() error = nil, want not nil")
		}
	}
	{
		var user User
		schema := NewSchema(
			Struct("user", &user, WithValidators(Required()), WithSubSchema(func(schema *Schema, user *User) {
				WithSchema(schema, Value("name", &user.Name, WithValidators(Required())))
				WithSchema(schema, Value("age", &user.Age, WithValidators(Required())))
			})),
		)

		jsonData := `{"user": {"name": "test", "age": 20}}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if user.Name != "test" {
			t.Errorf("Schema.ApplyJSON() user.Name = %v, want %v", user.Name, "test")
		}
		if user.Age != 20 {
			t.Errorf("Schema.ApplyJSON() user.Age = %v, want %v", user.Age, 20)
		}
	}
}

func TestPoxxy_Complex(t *testing.T) {
	type House struct {
		Address    string
		Price      int
		Rooms      []string
		Properties map[string]string
	}

	type User struct {
		House1 House
		House2 *House
	}

	var user User
	schema := NewSchema(
		Struct("user", &user, WithValidators(Required()), WithSubSchema(func(schema *Schema, user *User) {
			WithSchema(schema, Struct("house1", &user.House1, WithValidators(Required()), WithSubSchema(func(schema *Schema, house *House) {
				WithSchema(schema, Value("address", &house.Address, WithValidators(Required())))
				WithSchema(schema, Value("price", &house.Price, WithValidators(Required())))
				WithSchema(schema, Slice("rooms", &house.Rooms, WithValidators(Required())))
				WithSchema(schema, Map("properties", &house.Properties, WithValidators(Required(), WithMapKeys("color")), WithSubSchemaMap(func(ss *Schema, key string, value string) {
					if key == "color" {
						WithSchema(ss, ValueWithoutAssign[string](key, WithValidators(In("red", "blue", "green"))))
					}
				})))
			})))
			WithSchema(schema, Pointer("house2", &user.House2))
		})),
	)

	jsonData := `
	{
		"user": {
			"house1": {
				"address": "123 Main St",
				"price": 100000,
				"rooms": [
					"bedroom",
					"bathroom"
				],
				"properties": {
					"color": "red"
				}
			}
		}
	}
	`
	err := schema.ApplyJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Schema.ApplyJSON() error = %v", err)
	}

	if user.House1.Address != "123 Main St" {
		t.Errorf("Schema.ApplyJSON() user.House1.Address = %v, want %v", user.House1.Address, "123 Main St")
	}

	if user.House2 != nil {
		t.Errorf("Schema.ApplyJSON() user.House2 = %v, want %v", user.House2, nil)
	}

	if user.House1.Price != 100000 {
		t.Errorf("Schema.ApplyJSON() user.House1.Price = %v, want %v", user.House1.Price, 100000)
	}

	if len(user.House1.Rooms) != 2 {
		t.Errorf("Schema.ApplyJSON() user.House1.Rooms = %v, want %v", user.House1.Rooms, []string{"bedroom", "bathroom"})
	}

	if user.House1.Properties["color"] != "red" {
		t.Errorf("Schema.ApplyJSON() user.House1.Properties = %v, want %v", user.House1.Properties, map[string]string{"color": "red"})
	}
}

func TestPoxxy_SQLNullFields(t *testing.T) {
	type User struct {
		Name sql.NullString
		Age  sql.NullInt64
	}

	var user User
	schema := NewSchema(
		Value("name", &user.Name, WithValidators(Required())),
		Value("age", &user.Age, WithValidators(Required())),
	)

	jsonData := `{"name": "test", "age": 20}`
	err := schema.ApplyJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Schema.ApplyJSON() error = %v", err)
	}

	if !user.Name.Valid {
		t.Errorf("Schema.ApplyJSON() user.Name = %v, want %v", user.Name.String, "test")
	}

	if user.Name.String != "test" {
		t.Errorf("Schema.ApplyJSON() user.Name = %v, want %v", user.Name.String, "test")
	}

	if !user.Age.Valid {
		t.Errorf("Schema.ApplyJSON() user.Age = %v, want %v", user.Age.Int64, 20)
	}

	if user.Age.Int64 != 20 {
		t.Errorf("Schema.ApplyJSON() user.Age = %v, want %v", user.Age.Int64, 20)
	}
}
