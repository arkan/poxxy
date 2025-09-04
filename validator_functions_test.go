package poxxy

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequired(t *testing.T) {
	// Required validator now needs to be tested in schema context

	t.Run("field present should pass", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithValidators(Required())),
		)

		data := map[string]interface{}{
			"test": "hello", // Field is present
		}

		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Expected no error when field is present, got: %v", err)
		}
	})

	t.Run("field present with zero value should pass", func(t *testing.T) {
		var value int
		schema := NewSchema(
			Value("test", &value, WithValidators(Required())),
		)

		data := map[string]interface{}{
			"test": 0, // Field is present but zero value
		}

		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Expected no error when field is present with zero value, got: %v", err)
		}

		if value != 0 {
			t.Errorf("Expected value to be 0, got: %v", value)
		}
	})

	t.Run("field present with empty string should fail", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithValidators(Required())),
		)

		data := map[string]interface{}{
			"test": "", // Field is present but empty
		}

		err := schema.Apply(data)
		if err == nil {
			t.Error("Expected error when field is present with empty string, got nil")
		}

		if !strings.Contains(err.Error(), "field is required") {
			t.Errorf("Expected 'field is required' error, got: %v", err)
		}
	})

	t.Run("field missing should fail", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithValidators(Required())),
		)

		data := map[string]interface{}{
			// "test" field is missing
		}

		err := schema.Apply(data)
		if err == nil {
			t.Error("Expected error when field is missing, got nil")
		}

		if !strings.Contains(err.Error(), "field is required") {
			t.Errorf("Expected 'field is required' error, got: %v", err)
		}
	})

	t.Run("field present with '' value should fail", func(t *testing.T) {
		var value sql.NullString
		schema := NewSchema(
			Value("test", &value, WithValidators(Required())),
		)

		data := map[string]interface{}{
			"test": "",
		}

		err := schema.Apply(data)
		if assert.Error(t, err) {
			assert.Equal(t, "test: field is required", err.Error())
		}
	})
}

func TestEmail(t *testing.T) {
	validator := Email()

	// Valid email addresses
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@example.org",
		"123@numbers.com",
		"test-email@sub.domain.com",
		"",
	}

	for _, email := range validEmails {
		t.Run("valid: "+email, func(t *testing.T) {
			err := validator.Validate(email, "email")
			if err != nil {
				t.Errorf("Expected valid email %s to pass, got: %v", email, err)
			}
		})
	}

	// Invalid email addresses
	invalidEmails := []string{
		"invalid",
		"@example.com",
		"test@",
		"test.example.com",
		"test@.com",
		"test@com",
	}

	for _, email := range invalidEmails {
		t.Run("invalid: "+email, func(t *testing.T) {
			err := validator.Validate(email, "email")
			if err == nil {
				t.Errorf("Expected invalid email %s to fail, got no error", email)
			}
		})
	}

	// Test non-string input
	t.Run("non-string input", func(t *testing.T) {
		err := validator.Validate(123, "email")
		if err == nil {
			t.Error("Expected error for non-string input, got nil")
		}
	})
}

func TestMin(t *testing.T) {
	// Test with int
	validator := Min(10)

	t.Run("int above min", func(t *testing.T) {
		err := validator.Validate(15, "value")
		if err != nil {
			t.Errorf("Expected no error for value above min, got: %v", err)
		}
	})

	t.Run("int equal to min", func(t *testing.T) {
		err := validator.Validate(10, "value")
		if err != nil {
			t.Errorf("Expected no error for value equal to min, got: %v", err)
		}
	})

	t.Run("int below min", func(t *testing.T) {
		err := validator.Validate(5, "value")
		if err == nil {
			t.Error("Expected error for value below min, got nil")
		}
	})

	// Test with float64
	floatValidator := Min(10.5)

	t.Run("float above min", func(t *testing.T) {
		err := floatValidator.Validate(15.5, "value")
		if err != nil {
			t.Errorf("Expected no error for float above min, got: %v", err)
		}
	})

	t.Run("float below min", func(t *testing.T) {
		err := floatValidator.Validate(5.5, "value")
		if err == nil {
			t.Error("Expected error for float below min, got nil")
		}
	})

	t.Run("driver.Valuer", func(t *testing.T) {
		v := sql.NullFloat64{
			Float64: 50,
			Valid:   true,
		}
		validator := Min(float64(100))
		err := validator.Validate(v, "value")
		if err == nil {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
		if err.Error() != "value must be at least 100.000000" {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
	})

	t.Run("driver.Valuer with mismatched type (float64 and int)", func(t *testing.T) {
		v := sql.NullFloat64{
			Float64: 50,
			Valid:   true,
		}
		validator := Min(100)
		err := validator.Validate(v, "value")
		if err == nil {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
		if err.Error() != "value must be a int type" {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
	})
}

func TestMax(t *testing.T) {
	// Test with int
	validator := Max(100)

	t.Run("int below max", func(t *testing.T) {
		err := validator.Validate(50, "value")
		if err != nil {
			t.Errorf("Expected no error for value below max, got: %v", err)
		}
	})

	t.Run("int equal to max", func(t *testing.T) {
		err := validator.Validate(100, "value")
		if err != nil {
			t.Errorf("Expected no error for value equal to max, got: %v", err)
		}
	})

	t.Run("int above max", func(t *testing.T) {
		err := validator.Validate(150, "value")
		if err == nil {
			t.Error("Expected error for value above max, got nil")
		}
	})

	// Test with float64
	floatValidator := Max(100.5)

	t.Run("float below max", func(t *testing.T) {
		err := floatValidator.Validate(50.5, "value")
		if err != nil {
			t.Errorf("Expected no error for float below max, got: %v", err)
		}
	})

	t.Run("float above max", func(t *testing.T) {
		err := floatValidator.Validate(150.5, "value")
		if err == nil {
			t.Error("Expected error for float above max, got nil")
		}
	})

	t.Run("driver.Valuer", func(t *testing.T) {
		v := sql.NullInt64{
			Int64: 150,
			Valid: true,
		}
		validator := Max(int64(100))
		err := validator.Validate(v, "value")
		if err == nil {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
		if err.Error() != "value must be at most 100" {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
	})

	t.Run("driver.Valuer with mismatched type (float64 and int)", func(t *testing.T) {
		v := sql.NullFloat64{
			Float64: 100,
			Valid:   true,
		}
		validator := Max(50)
		err := validator.Validate(v, "value")
		if err == nil {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
		if err.Error() != "value must be a int type and not a float64 type" {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
	})
}

func TestMinLength(t *testing.T) {
	validator := MinLength(3)

	// Test with strings
	t.Run("string above min length", func(t *testing.T) {
		err := validator.Validate("hello", "value")
		if err != nil {
			t.Errorf("Expected no error for string above min length, got: %v", err)
		}
	})

	t.Run("string equal to min length", func(t *testing.T) {
		err := validator.Validate("abc", "value")
		if err != nil {
			t.Errorf("Expected no error for string equal to min length, got: %v", err)
		}
	})

	t.Run("string below min length", func(t *testing.T) {
		err := validator.Validate("ab", "value")
		if err == nil {
			t.Error("Expected error for string below min length, got nil")
		}
	})

	// Test with slices
	t.Run("slice above min length", func(t *testing.T) {
		err := validator.Validate([]string{"a", "b", "c", "d"}, "value")
		if err != nil {
			t.Errorf("Expected no error for slice above min length, got: %v", err)
		}
	})

	t.Run("slice below min length", func(t *testing.T) {
		err := validator.Validate([]string{"a", "b"}, "value")
		if err == nil {
			t.Error("Expected error for slice below min length, got nil")
		}
	})

	// Test with arrays
	t.Run("array above min length", func(t *testing.T) {
		err := validator.Validate([4]string{"a", "b", "c", "d"}, "value")
		if err != nil {
			t.Errorf("Expected no error for array above min length, got: %v", err)
		}
	})
}

func TestMaxLength(t *testing.T) {
	validator := MaxLength(5)

	// Test with strings
	t.Run("string below max length", func(t *testing.T) {
		err := validator.Validate("abc", "value")
		if err != nil {
			t.Errorf("Expected no error for string below max length, got: %v", err)
		}
	})

	t.Run("string equal to max length", func(t *testing.T) {
		err := validator.Validate("abcde", "value")
		if err != nil {
			t.Errorf("Expected no error for string equal to max length, got: %v", err)
		}
	})

	t.Run("string above max length", func(t *testing.T) {
		err := validator.Validate("abcdef", "value")
		if err == nil {
			t.Error("Expected error for string above max length, got nil")
		}
	})

	// Test with slices
	t.Run("slice below max length", func(t *testing.T) {
		err := validator.Validate([]string{"a", "b", "c"}, "value")
		if err != nil {
			t.Errorf("Expected no error for slice below max length, got: %v", err)
		}
	})

	t.Run("slice above max length", func(t *testing.T) {
		err := validator.Validate([]string{"a", "b", "c", "d", "e", "f"}, "value")
		if err == nil {
			t.Error("Expected error for slice above max length, got nil")
		}
	})
}

func TestURL(t *testing.T) {
	validator := URL()

	// Valid URLs
	validURLs := []string{
		"http://example.com",
		"https://example.com",
		"http://sub.domain.com/path",
		"https://example.com/path?query=value",
		"http://localhost:8080",
	}

	for _, url := range validURLs {
		t.Run("valid: "+url, func(t *testing.T) {
			err := validator.Validate(url, "url")
			if err != nil {
				t.Errorf("Expected valid URL %s to pass, got: %v", url, err)
			}
		})
	}

	// Invalid URLs
	invalidURLs := []string{
		"example.com",
		"ftp://example.com",
		"not-a-url",
		"http://",
		"https://",
	}

	for _, url := range invalidURLs {
		t.Run("invalid: "+url, func(t *testing.T) {
			err := validator.Validate(url, "url")
			if err == nil {
				t.Errorf("Expected invalid URL %s to fail, got no error", url)
			}
		})
	}

	// Test with empty string
	t.Run("empty string", func(t *testing.T) {
		err := validator.Validate("", "url")
		if err != nil {
			t.Errorf("Expected empty string to be valid, got: %v", err)
		}
	})

	// Test non-string input
	t.Run("non-string input", func(t *testing.T) {
		err := validator.Validate(123, "url")
		if err == nil {
			t.Error("Expected error for non-string input, got nil")
		}
	})
}

func TestIn(t *testing.T) {
	validator := In("apple", "banana", "orange")

	// Valid values
	validValues := []string{"apple", "banana", "orange"}
	for _, value := range validValues {
		t.Run("valid: "+value, func(t *testing.T) {
			err := validator.Validate(value, "fruit")
			if err != nil {
				t.Errorf("Expected value %s to be valid, got: %v", value, err)
			}
		})
	}

	// Invalid values
	invalidValues := []string{"grape", "kiwi", ""}
	for _, value := range invalidValues {
		t.Run("invalid: "+value, func(t *testing.T) {
			err := validator.Validate(value, "fruit")
			if err == nil {
				t.Errorf("Expected value %s to be invalid, got no error", value)
			}
		})
	}

	// Test with different types
	numValidator := In(1, 2, 3)

	t.Run("valid number", func(t *testing.T) {
		err := numValidator.Validate(2, "number")
		if err != nil {
			t.Errorf("Expected number 2 to be valid, got: %v", err)
		}
	})

	t.Run("invalid number", func(t *testing.T) {
		err := numValidator.Validate(4, "number")
		if err == nil {
			t.Error("Expected number 4 to be invalid, got no error")
		}
	})

	t.Run("driver.Valuer", func(t *testing.T) {
		v := sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		validator := In(int64(1), int64(2), int64(3))
		err := validator.Validate(v, "number")
		if err != nil {
			t.Errorf("Expected no error for driver.Valuer, got: %v", err)
		}
	})

	t.Run("driver.Valuer with mismatched type (float64 and int)", func(t *testing.T) {
		v := sql.NullFloat64{
			Float64: 2.5,
			Valid:   true,
		}
		validator := In(1, 2, 3)
		err := validator.Validate(v, "number")
		if err == nil {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
		if err.Error() != "value must be one of: [1 2 3]" {
			t.Errorf("Expected error for driver.Valuer, got: %v", err)
		}
	})
}

func TestEach(t *testing.T) {
	// Create validators for each element
	elementValidator := Each(Required(), MinLength(2))

	// Test with valid slice
	t.Run("valid slice", func(t *testing.T) {
		validSlice := []string{"hello", "world", "test"}
		err := elementValidator.Validate(validSlice, "items")
		if err != nil {
			t.Errorf("Expected valid slice to pass, got: %v", err)
		}
	})

	// Test with invalid slice (empty string)
	t.Run("invalid slice with empty string", func(t *testing.T) {
		invalidSlice := []string{"hello", "", "test"}
		err := elementValidator.Validate(invalidSlice, "items")
		if err == nil {
			t.Error("Expected slice with empty string to fail, got no error")
		}
	})

	// Test with invalid slice (short string)
	t.Run("invalid slice with short string", func(t *testing.T) {
		invalidSlice := []string{"hello", "a", "test"}
		err := elementValidator.Validate(invalidSlice, "items")
		if err == nil {
			t.Error("Expected slice with short string to fail, got no error")
		}
	})

	// Test with array
	t.Run("valid array", func(t *testing.T) {
		validArray := [3]string{"hello", "world", "test"}
		err := elementValidator.Validate(validArray, "items")
		if err != nil {
			t.Errorf("Expected valid array to pass, got: %v", err)
		}
	})

	// Test with non-slice/array
	t.Run("non-slice input", func(t *testing.T) {
		err := elementValidator.Validate("not a slice", "items")
		if err == nil {
			t.Error("Expected error for non-slice input, got nil")
		}
	})
}

func TestUnique(t *testing.T) {
	validator := Unique()

	// Test with slice without duplicates
	t.Run("slice without duplicates", func(t *testing.T) {
		slice := []string{"apple", "banana", "orange"}
		err := validator.Validate(slice, "fruits")
		if err != nil {
			t.Errorf("Expected slice without duplicates to pass, got: %v", err)
		}
	})

	// Test with slice with duplicates
	t.Run("slice with duplicates", func(t *testing.T) {
		slice := []string{"apple", "banana", "apple"}
		err := validator.Validate(slice, "fruits")
		if err == nil {
			t.Error("Expected slice with duplicates to fail, got no error")
		}
	})

	// Test with array without duplicates
	t.Run("array without duplicates", func(t *testing.T) {
		array := [3]int{1, 2, 3}
		err := validator.Validate(array, "numbers")
		if err != nil {
			t.Errorf("Expected array without duplicates to pass, got: %v", err)
		}
	})

	// Test with array with duplicates
	t.Run("array with duplicates", func(t *testing.T) {
		array := [3]int{1, 2, 1}
		err := validator.Validate(array, "numbers")
		if err == nil {
			t.Error("Expected array with duplicates to fail, got no error")
		}
	})

	// Test with map without duplicate values
	t.Run("map without duplicate values", func(t *testing.T) {
		m := map[string]string{"a": "apple", "b": "banana", "c": "orange"}
		err := validator.Validate(m, "mapping")
		if err != nil {
			t.Errorf("Expected map without duplicate values to pass, got: %v", err)
		}
	})

	// Test with map with duplicate values
	t.Run("map with duplicate values", func(t *testing.T) {
		m := map[string]string{"a": "apple", "b": "banana", "c": "apple"}
		err := validator.Validate(m, "mapping")
		if err == nil {
			t.Error("Expected map with duplicate values to fail, got no error")
		}
	})

	// Test with unsupported type
	t.Run("unsupported type", func(t *testing.T) {
		err := validator.Validate("not a collection", "value")
		if err == nil {
			t.Error("Expected error for unsupported type, got nil")
		}
	})

	// Test with empty collections
	t.Run("empty slice", func(t *testing.T) {
		slice := []string{}
		err := validator.Validate(slice, "empty")
		if err != nil {
			t.Errorf("Expected empty slice to pass, got: %v", err)
		}
	})
}

func TestUniqueBy(t *testing.T) {
	type Person struct {
		ID   int
		Name string
	}

	// Key extractor function
	keyExtractor := func(item interface{}) interface{} {
		if person, ok := item.(Person); ok {
			return person.ID
		}
		return nil
	}

	validator := UniqueBy(keyExtractor)

	// Test with slice without duplicate IDs
	t.Run("slice without duplicate keys", func(t *testing.T) {
		people := []Person{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
			{ID: 3, Name: "Charlie"},
		}
		err := validator.Validate(people, "people")
		if err != nil {
			t.Errorf("Expected slice without duplicate keys to pass, got: %v", err)
		}
	})

	// Test with slice with duplicate IDs
	t.Run("slice with duplicate keys", func(t *testing.T) {
		people := []Person{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
			{ID: 1, Name: "Alice2"}, // Duplicate ID
		}
		err := validator.Validate(people, "people")
		if err == nil {
			t.Error("Expected slice with duplicate keys to fail, got no error")
		}
	})

	// Test with array
	t.Run("array without duplicate keys", func(t *testing.T) {
		people := [2]Person{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
		}
		err := validator.Validate(people, "people")
		if err != nil {
			t.Errorf("Expected array without duplicate keys to pass, got: %v", err)
		}
	})

	// Test with unsupported type
	t.Run("unsupported type", func(t *testing.T) {
		err := validator.Validate("not a collection", "value")
		if err == nil {
			t.Error("Expected error for unsupported type, got nil")
		}
	})

	// Test with empty slice
	t.Run("empty slice", func(t *testing.T) {
		people := []Person{}
		err := validator.Validate(people, "empty")
		if err != nil {
			t.Errorf("Expected empty slice to pass, got: %v", err)
		}
	})
}

func TestNotEmpty(t *testing.T) {
	validator := NotEmpty()

	// Test cases that should pass
	testCases := []struct {
		name  string
		value interface{}
	}{
		{"non-empty string", "hello"},
		{"non-zero int", 42},
		{"non-zero float", 3.14},
		{"true bool", true},
		{"false bool", false},
		{"non-empty slice", []string{"item"}},
		{"non-empty map", map[string]string{"key": "value"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" should pass", func(t *testing.T) {
			err := validator.Validate(tc.value, "testField")
			assert.NoError(t, err)
		})
	}

	// Test cases that should fail
	failCases := []struct {
		name  string
		value interface{}
	}{
		{"nil value", nil},
		{"empty string", ""},
		// {"zero int", 0},
		// {"zero uint", uint(0)},
		// {"zero float", 0.0},
		{"empty slice", []string{}},
		{"empty map", map[string]string{}},
	}

	for _, tc := range failCases {
		t.Run(tc.name+" should fail", func(t *testing.T) {
			err := validator.Validate(tc.value, "testField")
			assert.Error(t, err)
		})
	}
}

func TestValidatorWithMessage(t *testing.T) {
	t.Run("Required custom error message", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithValidators(Required().WithMessage("Custom required message"))),
		)

		data := map[string]interface{}{
			// "test" field is missing
		}

		err := schema.Apply(data)
		assert.Error(t, err)
		assert.Equal(t, "test: Custom required message", err.Error())
	})

	t.Run("NonZero custom error message", func(t *testing.T) {
		validator := NotEmpty().WithMessage("Custom zero message")
		err := validator.Validate("", "field")
		assert.Error(t, err)
		assert.Equal(t, "Custom zero message", err.Error())
	})
}

func TestSlice(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	// Test successful assignment and validation
	t.Run("valid slice of structs", func(t *testing.T) {
		var people []Person

		schema := NewSchema(
			Slice[Person]("people", &people,
				WithSubSchema(func(s *Schema, p *Person) {
					WithSchema(s, Value[string]("name", &p.Name, WithValidators(Required())))
					WithSchema(s, Value[int]("age", &p.Age, WithValidators(Min(0))))
				}),
				WithValidators(Required()),
			),
		)

		data := map[string]interface{}{
			"people": []map[string]interface{}{
				{"name": "Alice", "age": 25},
				{"name": "Bob", "age": 30},
			},
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(people))

		assert.Equal(t, "Alice", people[0].Name)
		assert.Equal(t, 25, people[0].Age)
		assert.Equal(t, "Bob", people[1].Name)
		assert.Equal(t, 30, people[1].Age)
	})

	// Test validation failure on element
	t.Run("invalid element in slice", func(t *testing.T) {
		var people []Person

		schema := NewSchema(
			Slice[Person]("people", &people,
				WithSubSchema(func(s *Schema, p *Person) {
					WithSchema(s, Value[string]("name", &p.Name, WithValidators(Required())))
					WithSchema(s, Value[int]("age", &p.Age, WithValidators(Min(0))))
				}),
				WithValidators(MinLength(2)),
			),
		)

		data := map[string]interface{}{
			"people": []map[string]interface{}{
				{"name": "Alice", "age": 25},
				{"name": "", "age": -5}, // Invalid: empty name and negative age
			},
		}

		err := schema.Apply(data)
		assert.Error(t, err)
		assert.Equal(t, "people: element 1: name: field is required; age: value must be at least 0; people: must have at least 2 items", err.Error())
	})

	// Test slice-level validation
	t.Run("slice level validation", func(t *testing.T) {
		var people []Person

		schema := NewSchema(
			Slice[Person]("people", &people,
				WithSubSchema(func(s *Schema, p *Person) {
					WithSchema(s, Value[string]("name", &p.Name, WithValidators(Required())))
					WithSchema(s, Value[int]("age", &p.Age, WithValidators(Min(0))))
				}),
				WithValidators(MinLength(2)),
			),
		)

		data := map[string]interface{}{
			"people": []map[string]interface{}{
				{"name": "Alice", "age": 25},
			},
		}

		err := schema.Apply(data)
		assert.Error(t, err)
		assert.Equal(t, "people: must have at least 2 items", err.Error())
	})

	// Test with []interface{} input
	t.Run("slice with interface{} elements", func(t *testing.T) {
		var people []Person

		schema := NewSchema(
			Slice[Person]("people", &people,
				WithSubSchema(func(s *Schema, p *Person) {
					WithSchema(s, Value[string]("name", &p.Name, WithValidators(Required())))
					WithSchema(s, Value[int]("age", &p.Age, WithValidators(Min(0))))
				}),
			),
		)

		data := map[string]interface{}{
			"people": []interface{}{
				map[string]interface{}{"name": "Alice", "age": 25},
				map[string]interface{}{"name": "Bob", "age": 30},
			},
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(people))
	})

	// Test empty slice
	t.Run("empty slice", func(t *testing.T) {
		var people []Person

		schema := NewSchema(
			Slice[Person]("people", &people,
				WithSubSchema(func(s *Schema, p *Person) {
					WithSchema(s, Value[string]("name", &p.Name, WithValidators(Required())))
					WithSchema(s, Value[int]("age", &p.Age, WithValidators(Min(0))))
				}),
			),
		)

		data := map[string]interface{}{
			"people": []map[string]interface{}{},
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(people))
	})
}
