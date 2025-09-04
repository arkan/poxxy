package poxxy

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, "test", name)
	assert.Equal(t, 20, age)
	assert.Equal(t, true, isAdmin)
	assert.Equal(t, []string{"tag1", "tag2"}, tags)
	assert.NotNil(t, label)
	assert.Equal(t, "okay", *label)
	assert.Nil(t, label2)
	assert.NotNil(t, label3)
	assert.Equal(t, int64(0), *label3)
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
	assert.NoError(t, err)

	assert.Equal(t, "test", user.Name)
	assert.Equal(t, 20, user.Age)
	expectedPreferences := map[string]string{"color": "blue", "size": "large"}
	assert.Equal(t, expectedPreferences, user.Preferences)
}

func TestPoxxy_Map_DefaultValue(t *testing.T) {
	// Test default value when field is missing
	t.Run("default value applied when field missing", func(t *testing.T) {
		var userSettings map[string]string
		defaultSettings := map[string]string{
			"theme":    "dark",
			"language": "en",
			"timezone": "UTC",
		}

		schema := NewSchema(
			Map("settings", &userSettings,
				WithDefault(defaultSettings),
			),
		)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.NoError(t, err)

		assert.Equal(t, 3, len(userSettings))
		assert.Equal(t, "dark", userSettings["theme"])
		assert.Equal(t, "en", userSettings["language"])
		assert.Equal(t, "UTC", userSettings["timezone"])
	})

	// Test that provided value overrides default
	t.Run("provided value overrides default", func(t *testing.T) {
		var userSettings map[string]string
		defaultSettings := map[string]string{
			"theme":    "dark",
			"language": "en",
			"timezone": "UTC",
		}

		schema := NewSchema(
			Map("settings", &userSettings,
				WithDefault(defaultSettings),
			),
		)

		// Apply JSON with settings data - should override default
		jsonData := `{"settings": {"theme": "light", "language": "es", "notifications": "on"}}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.NoError(t, err)

		assert.Equal(t, 3, len(userSettings))
		assert.Equal(t, "light", userSettings["theme"])
		assert.Equal(t, "es", userSettings["language"])
		assert.Equal(t, "on", userSettings["notifications"])
		assert.Equal(t, "", userSettings["timezone"]) // It must be empty.
	})

	// Test default value with nil field
	t.Run("default value applied when field is nil", func(t *testing.T) {
		var userSettings map[string]string
		defaultSettings := map[string]string{
			"theme":    "dark",
			"language": "en",
			"timezone": "UTC",
		}

		schema := NewSchema(
			Map("settings", &userSettings,
				WithDefault(defaultSettings),
			),
		)

		// Apply JSON with nil settings - should NOT use default value when explicitly null
		jsonData := `{"settings": null}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.NoError(t, err)

		// When field is explicitly null, default value should not be applied
		// The field should remain unassigned (nil map)
		if assert.NotNil(t, userSettings) {
			assert.Equal(t, 3, len(userSettings))
			assert.Equal(t, "dark", userSettings["theme"])
			assert.Equal(t, "en", userSettings["language"])
			assert.Equal(t, "UTC", userSettings["timezone"])
		}
	})

	// Test SetDefaultValue method directly
	t.Run("SetDefaultValue method works", func(t *testing.T) {
		var userSettings map[string]string
		defaultSettings := map[string]string{
			"theme":    "method_default",
			"language": "fr",
			"timezone": "EST",
		}

		// Create map field directly to test SetDefaultValue method
		mapField := &MapField[string, string]{
			name: "settings",
			ptr:  &userSettings,
		}

		// Set default value directly
		mapField.SetDefaultValue(defaultSettings)

		// Create schema with the field
		schema := NewSchema(mapField)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.NoError(t, err)

		assert.Equal(t, 3, len(userSettings))
		assert.Equal(t, "method_default", userSettings["theme"])
		assert.Equal(t, "fr", userSettings["language"])
		assert.Equal(t, "EST", userSettings["timezone"])
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
			Map("scores", &userScores,
				WithDefault(defaultScores),
			),
		)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.NoError(t, err)
		assert.Equal(t, 3, len(userScores))
		assert.Equal(t, 95.5, userScores[1])
		assert.Equal(t, 87.2, userScores[2])
		assert.Equal(t, 92.8, userScores[3])
	})

	// Test with callback functionality
	t.Run("works with callback functionality", func(t *testing.T) {
		var userSettings map[string]string
		defaultSettings := map[string]string{
			"theme":    "dark",
			"language": "en",
		}

		schema := NewSchema(
			Map("settings", &userSettings,
				WithDefault(defaultSettings),
				WithSubSchemaMap(func(s *Schema, k string, v string) {
					// Optional validation for individual map entries
				}),
			),
		)

		// Apply empty JSON - should use default value
		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.NoError(t, err)

		assert.Equal(t, 2, len(userSettings))
		assert.Equal(t, "dark", userSettings["theme"])
		assert.Equal(t, "en", userSettings["language"])
	})
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
	assert.NoError(t, err)

	expectedUsers := []User{{Name: "test", Age: 20}, {Name: "test2", Age: 21}}
	assert.Equal(t, expectedUsers, users)
	assert.Equal(t, 20, users[0].Age)
	assert.Equal(t, "test2", users[1].Name)
	assert.Equal(t, 21, users[1].Age)
}

func TestPoxxy_Boolean(t *testing.T) {
	{
		var isAdmin bool
		schema := NewSchema(
			Value("isAdmin", &isAdmin, WithValidators(Required())),
		)

		jsonData := `{"isAdmin": true}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.NoError(t, err)
		assert.True(t, isAdmin)
	}

	{
		var isAdmin *bool
		schema := NewSchema(
			Pointer("isAdmin", &isAdmin, WithValidators(Required())),
		)

		jsonData := `{"isAdmin": true}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.NoError(t, err)
		assert.NotNil(t, isAdmin)
		assert.True(t, *isAdmin)
	}

	{
		var isAdmin *bool
		schema := NewSchema(
			Pointer("isAdmin", &isAdmin, WithValidators(Required())),
		)

		jsonData := `{}`
		err := schema.ApplyJSON([]byte(jsonData))
		assert.Error(t, err)
		assert.Equal(t, "isAdmin: field is required", err.Error())
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
		assert.Error(t, err)
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
		assert.NoError(t, err)
		assert.Equal(t, "test", user.Name)
		assert.Equal(t, 20, user.Age)
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
	assert.NoError(t, err)
	assert.Equal(t, "123 Main St", user.House1.Address)
	assert.Nil(t, user.House2)
	assert.Equal(t, 100000, user.House1.Price)
	assert.Equal(t, 2, len(user.House1.Rooms))
	assert.Equal(t, "red", user.House1.Properties["color"])
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
	assert.NoError(t, err)
	assert.True(t, user.Name.Valid)
	assert.Equal(t, "test", user.Name.String)
	assert.True(t, user.Age.Valid)
	assert.Equal(t, int64(20), user.Age.Int64)
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
		assert.NoError(t, err)
		assert.Equal(t, "Default User", user.Name)
		assert.Equal(t, 25, user.Age)
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
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, 30, user.Age)
	})

	// Test default value with nil field
	t.Run("default value applied when field is nil", func(t *testing.T) {
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
		assert.NoError(t, err)

		// When field is explicitly null, default value should not be applied
		// The field should remain unassigned (zero values)
		assert.Equal(t, "Default User", user.Name)
		assert.Equal(t, 25, user.Age)
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
		assert.NoError(t, err)
		assert.Equal(t, "Method Default", user.Name)
		assert.Equal(t, 42, user.Age)
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
		assert.NoError(t, err)
		assert.Equal(t, 3, len(userPreferences))
		assert.Equal(t, "dark", userPreferences["theme"])
		assert.Equal(t, "en", userPreferences["language"])
		assert.Equal(t, "UTC", userPreferences["timezone"])
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
		assert.NoError(t, err)
		assert.Equal(t, 3, len(userPreferences))
		assert.Equal(t, "light", userPreferences["theme"])
		assert.Equal(t, "es", userPreferences["language"])
		assert.Equal(t, "on", userPreferences["notifications"])
		// Default value should not be present
		assert.Empty(t, userPreferences["timezone"])
	})

	// Test default value with nil field
	t.Run("default value applied when field is nil", func(t *testing.T) {
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
		assert.NoError(t, err)

		// When field is explicitly null, default value should not be applied
		// The field should remain unassigned (nil map)
		if assert.NotNil(t, userPreferences) {
			assert.Equal(t, 3, len(userPreferences))
			assert.Equal(t, "dark", userPreferences["theme"])
			assert.Equal(t, "en", userPreferences["language"])
			assert.Equal(t, "UTC", userPreferences["timezone"])
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
		assert.NoError(t, err)

		assert.Equal(t, 3, len(userPreferences))
		assert.Equal(t, "method_default", userPreferences["theme"])
		assert.Equal(t, "fr", userPreferences["language"])
		assert.Equal(t, "EST", userPreferences["timezone"])
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
		assert.NoError(t, err)

		assert.Equal(t, 3, len(userScores))
		assert.Equal(t, 95.5, userScores[1])
		assert.Equal(t, 87.2, userScores[2])
		assert.Equal(t, 92.8, userScores[3])
	})
}
