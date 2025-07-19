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
	var label *string
	var label2 *int64
	var label3 *int64

	schema := NewSchema(
		Value("name", &name, WithValidators(Required())),
		Value("age", &age, WithValidators(Required())),
		Value("isAdmin", &isAdmin, WithValidators(Required())),
		Slice("tags", &tags, WithValidators(Required())),
		Pointer("label", &label, WithValidators(Required())),
		Pointer("label2", &label2),
		Pointer("label3", &label3),
	)

	jsonData := `{"name": "test", "age": 20, "isAdmin": true, "tags": ["tag1", "tag2"], "label": "okay", "label2": "", "label3": "0"}`
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
	if label == nil || *label != "okay" {
		t.Errorf("Schema.ApplyJSON() label = %v, want %v", label, "okay")
	}
	if label2 != nil {
		t.Errorf("Schema.ApplyJSON() label2 = %v, want %v", label2, nil)
	}
	if label3 == nil || *label3 != 0 {
		t.Errorf("Schema.ApplyJSON() label3 = %v, want %v", label3, 0)
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

func TestPoxxy_Struct_DefaultValue(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	// Test default value when field is missing
	t.Run("default value applied when field missing", func(t *testing.T) {
		var user User
		defaultUser := User{Name: "Default User", Age: 25}

		schema := NewSchema(
			Struct("user", &user,
				WithDefault(defaultUser),
				WithSubSchema(func(schema *Schema, user *User) {
					WithSchema(schema, Value("name", &user.Name, WithValidators(Required())))
					WithSchema(schema, Value("age", &user.Age, WithValidators(Required())))
				}),
			),
		)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if user.Name != "Default User" {
			t.Errorf("Schema.ApplyJSON() user.Name = %v, want %v", user.Name, "Default User")
		}
		if user.Age != 25 {
			t.Errorf("Schema.ApplyJSON() user.Age = %v, want %v", user.Age, 25)
		}
	})

	// Test that provided value overrides default
	t.Run("provided value overrides default", func(t *testing.T) {
		var user User
		defaultUser := User{Name: "Default User", Age: 25}

		schema := NewSchema(
			Struct("user", &user,
				WithDefault(defaultUser),
				WithSubSchema(func(schema *Schema, user *User) {
					WithSchema(schema, Value("name", &user.Name, WithValidators(Required())))
					WithSchema(schema, Value("age", &user.Age, WithValidators(Required())))
				}),
			),
		)

		// Apply JSON with user data - should override default
		jsonData := `{"user": {"name": "John Doe", "age": 30}}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if user.Name != "John Doe" {
			t.Errorf("Schema.ApplyJSON() user.Name = %v, want %v", user.Name, "John Doe")
		}
		if user.Age != 30 {
			t.Errorf("Schema.ApplyJSON() user.Age = %v, want %v", user.Age, 30)
		}
	})

	// Test default value with nil field
	t.Run("default value not applied when field is explicitly nil", func(t *testing.T) {
		var user User
		defaultUser := User{Name: "Default User", Age: 25}

		schema := NewSchema(
			Struct("user", &user,
				WithDefault(defaultUser),
				WithSubSchema(func(schema *Schema, user *User) {
					WithSchema(schema, Value("name", &user.Name, WithValidators(Required())))
					WithSchema(schema, Value("age", &user.Age, WithValidators(Required())))
				}),
			),
		)

		// Apply JSON with nil user - should NOT use default value when explicitly null
		jsonData := `{"user": null}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		// When field is explicitly null, default value should not be applied
		// The field should remain unassigned (zero values)
		if user.Name != "" {
			t.Errorf("Schema.ApplyJSON() user.Name = %v, want empty string", user.Name)
		}
		if user.Age != 0 {
			t.Errorf("Schema.ApplyJSON() user.Age = %v, want 0", user.Age)
		}
	})

	// Test SetDefaultValue method directly
	t.Run("SetDefaultValue method works", func(t *testing.T) {
		var user User
		defaultUser := User{Name: "Method Default", Age: 42}

		// Create struct field directly to test SetDefaultValue method
		structField := &StructField[User]{
			name: "user",
			ptr:  &user,
		}

		// Set the callback
		structField.SetCallback(func(schema *Schema, user *User) {
			WithSchema(schema, Value("name", &user.Name, WithValidators(Required())))
			WithSchema(schema, Value("age", &user.Age, WithValidators(Required())))
		})

		// Set default value directly
		structField.SetDefaultValue(defaultUser)

		// Create schema with the field
		schema := NewSchema(structField)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if user.Name != "Method Default" {
			t.Errorf("Schema.ApplyJSON() user.Name = %v, want %v", user.Name, "Method Default")
		}
		if user.Age != 42 {
			t.Errorf("Schema.ApplyJSON() user.Age = %v, want %v", user.Age, 42)
		}
	})
}

func TestPoxxy_NestedMap_DefaultValue(t *testing.T) {
	// Test default value when field is missing
	t.Run("default value applied when field missing", func(t *testing.T) {
		var userPreferences map[string]string
		defaultPreferences := map[string]string{
			"theme":    "dark",
			"language": "en",
			"timezone": "UTC",
		}

		schema := NewSchema(
			NestedMap("preferences", &userPreferences,
				WithDefault(defaultPreferences),
				WithSubSchemaMap(func(s *Schema, k string, v string) {
					// Optional validation for individual map entries
				}),
			),
		)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if len(userPreferences) != 3 {
			t.Errorf("Schema.ApplyJSON() userPreferences length = %v, want %v", len(userPreferences), 3)
		}
		if userPreferences["theme"] != "dark" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"theme\"] = %v, want %v", userPreferences["theme"], "dark")
		}
		if userPreferences["language"] != "en" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"language\"] = %v, want %v", userPreferences["language"], "en")
		}
		if userPreferences["timezone"] != "UTC" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"timezone\"] = %v, want %v", userPreferences["timezone"], "UTC")
		}
	})

	// Test that provided value overrides default
	t.Run("provided value overrides default", func(t *testing.T) {
		var userPreferences map[string]string
		defaultPreferences := map[string]string{
			"theme":    "dark",
			"language": "en",
			"timezone": "UTC",
		}

		schema := NewSchema(
			NestedMap("preferences", &userPreferences,
				WithDefault(defaultPreferences),
				WithSubSchemaMap(func(s *Schema, k string, v string) {
					// Optional validation for individual map entries
				}),
			),
		)

		// Apply JSON with preferences data - should override default
		jsonData := `{"preferences": {"theme": "light", "language": "es", "notifications": "on"}}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if len(userPreferences) != 3 {
			t.Errorf("Schema.ApplyJSON() userPreferences length = %v, want %v", len(userPreferences), 3)
		}
		if userPreferences["theme"] != "light" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"theme\"] = %v, want %v", userPreferences["theme"], "light")
		}
		if userPreferences["language"] != "es" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"language\"] = %v, want %v", userPreferences["language"], "es")
		}
		if userPreferences["notifications"] != "on" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"notifications\"] = %v, want %v", userPreferences["notifications"], "on")
		}
		// Default value should not be present
		if _, exists := userPreferences["timezone"]; exists {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"timezone\"] should not exist, but got %v", userPreferences["timezone"])
		}
	})

	// Test default value with nil field
	t.Run("default value not applied when field is explicitly nil", func(t *testing.T) {
		var userPreferences map[string]string
		defaultPreferences := map[string]string{
			"theme":    "dark",
			"language": "en",
			"timezone": "UTC",
		}

		schema := NewSchema(
			NestedMap("preferences", &userPreferences,
				WithDefault(defaultPreferences),
				WithSubSchemaMap(func(s *Schema, k string, v string) {
					// Optional validation for individual map entries
				}),
			),
		)

		// Apply JSON with nil preferences - should NOT use default value when explicitly null
		jsonData := `{"preferences": null}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		// When field is explicitly null, default value should not be applied
		// The field should remain unassigned (nil map)
		if userPreferences != nil {
			t.Errorf("Schema.ApplyJSON() userPreferences = %v, want nil", userPreferences)
		}
	})

	// Test SetDefaultValue method directly
	t.Run("SetDefaultValue method works", func(t *testing.T) {
		var userPreferences map[string]string
		defaultPreferences := map[string]string{
			"theme":    "method_default",
			"language": "fr",
			"timezone": "EST",
		}

		// Create nested map field directly to test SetDefaultValue method
		nestedMapField := &NestedMapField[string, string]{
			name: "preferences",
			ptr:  &userPreferences,
		}

		// Set the callback
		nestedMapField.SetCallback(func(s *Schema, k string, v string) {
			// Optional validation for individual map entries
		})

		// Set default value directly
		nestedMapField.SetDefaultValue(defaultPreferences)

		// Create schema with the field
		schema := NewSchema(nestedMapField)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if len(userPreferences) != 3 {
			t.Errorf("Schema.ApplyJSON() userPreferences length = %v, want %v", len(userPreferences), 3)
		}
		if userPreferences["theme"] != "method_default" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"theme\"] = %v, want %v", userPreferences["theme"], "method_default")
		}
		if userPreferences["language"] != "fr" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"language\"] = %v, want %v", userPreferences["language"], "fr")
		}
		if userPreferences["timezone"] != "EST" {
			t.Errorf("Schema.ApplyJSON() userPreferences[\"timezone\"] = %v, want %v", userPreferences["timezone"], "EST")
		}
	})

	// Test with different key and value types
	t.Run("works with different key and value types", func(t *testing.T) {
		var userScores map[int]float64
		defaultScores := map[int]float64{
			1: 95.5,
			2: 87.2,
			3: 92.8,
		}

		schema := NewSchema(
			NestedMap("scores", &userScores,
				WithDefault(defaultScores),
				WithSubSchemaMap(func(s *Schema, k int, v float64) {
					// Optional validation for individual map entries
				}),
			),
		)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		if err != nil {
			t.Errorf("Schema.ApplyJSON() error = %v", err)
		}

		if len(userScores) != 3 {
			t.Errorf("Schema.ApplyJSON() userScores length = %v, want %v", len(userScores), 3)
		}
		if userScores[1] != 95.5 {
			t.Errorf("Schema.ApplyJSON() userScores[1] = %v, want %v", userScores[1], 95.5)
		}
		if userScores[2] != 87.2 {
			t.Errorf("Schema.ApplyJSON() userScores[2] = %v, want %v", userScores[2], 87.2)
		}
		if userScores[3] != 92.8 {
			t.Errorf("Schema.ApplyJSON() userScores[3] = %v, want %v", userScores[3], 92.8)
		}
	})
}
