package poxxy

import (
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConversion_EdgeCases(t *testing.T) {
	t.Run("string to numeric overflow", func(t *testing.T) {
		var intValue int
		var int64Value int64
		var floatValue float64

		schema := NewSchema(
			Value("int", &intValue),
			Value("int64", &int64Value),
			Value("float", &floatValue),
		)

		// Test with values that exceed limits
		overflowCases := []struct {
			field      string
			value      string
			shouldPass bool
		}{
			{"int", "9223372036854775808", false},    // Max int64 + 1
			{"int", "-9223372036854775809", false},   // Min int64 - 1
			{"int64", "9223372036854775808", false},  // Max int64 + 1
			{"int64", "-9223372036854775809", false}, // Min int64 - 1
			{"float", "1e309", false},                // Greater than MaxFloat64
			{"float", "-1e309", false},               // Less than -MaxFloat64
		}

		for _, tc := range overflowCases {
			err := schema.Apply(map[string]interface{}{tc.field: tc.value})
			if tc.shouldPass {
				assert.NoError(t, err, "Should pass for %s: %s", tc.field, tc.value)
			} else {
				assert.Error(t, err, "Should fail for %s: %s", tc.field, tc.value)
			}
		}
	})

	t.Run("invalid numeric formats", func(t *testing.T) {
		var intValue int
		var floatValue float64

		schema := NewSchema(
			Value("int", &intValue),
			Value("float", &floatValue),
		)

		invalidCases := []struct {
			field string
			value string
		}{
			{"int", "12.34.56"},
			{"int", "abc"},
			{"int", "12abc"},
			{"int", "abc12"},
			{"float", "12.34.56"},
			{"float", "abc"},
			{"float", "12abc"},
			{"float", "abc12"},
		}

		for _, tc := range invalidCases {
			err := schema.Apply(map[string]interface{}{tc.field: tc.value})
			assert.Error(t, err, "Should fail for %s: %s", tc.field, tc.value)
		}
	})

	t.Run("bool conversion edge cases", func(t *testing.T) {
		var boolValue bool

		schema := NewSchema(
			Value("bool", &boolValue),
		)

		boolCases := []struct {
			input      interface{}
			expected   bool
			shouldPass bool
		}{
			{"true", true, true},
			{"TRUE", true, true},
			{"True", true, true},
			{"false", false, true},
			{"FALSE", false, true},
			{"False", false, true},
			{"1", true, true},
			{"0", false, true},
			{"yes", true, true},
			{"no", false, true},
			{"y", true, true},
			{"n", false, true},
			{"on", true, true},
			{"off", false, true},
			{"t", true, true},
			{"f", false, true},
			{"maybe", false, false},
			{"invalid", false, false},
			{"", false, true}, // go-convert converts empty strings to false
		}

		for _, tc := range boolCases {
			err := schema.Apply(map[string]interface{}{"bool": tc.input})
			if tc.shouldPass {
				assert.NoError(t, err, "Should pass for %v", tc.input)
				assert.Equal(t, tc.expected, boolValue, "Wrong value for %v", tc.input)
			} else {
				assert.Error(t, err, "Should fail for %v", tc.input)
			}
		}
	})

	t.Run("nil to non-pointer conversion", func(t *testing.T) {
		var stringValue string
		var intValue int
		var boolValue bool

		schema := NewSchema(
			Value("string", &stringValue),
			Value("int", &intValue),
			Value("bool", &boolValue),
		)

		// Test with nil values - go-convert converts nil to zero values
		err := schema.Apply(map[string]interface{}{
			"string": nil,
			"int":    nil,
			"bool":   nil,
		})
		assert.NoError(t, err)
		assert.Equal(t, "", stringValue)
		assert.Equal(t, 0, intValue)
		assert.Equal(t, false, boolValue)
	})

	t.Run("complex type conversion", func(t *testing.T) {
		var stringValue string
		var intValue int

		schema := NewSchema(
			Value("string", &stringValue),
			Value("int", &intValue),
		)

		complexCases := []struct {
			field      string
			value      interface{}
			shouldPass bool
		}{
			{"string", map[string]interface{}{"key": "value"}, true}, // go-convert can convert map to string
			{"string", []interface{}{"item"}, true},                  // go-convert can convert slice to string
			{"string", func() {}, true},                              // go-convert can convert functions to string
			{"string", make(chan int), true},                         // go-convert can convert channels to string
			{"int", map[string]interface{}{"key": "value"}, false},   // map to int fails
			{"int", []interface{}{1, 2, 3}, false},                   // slice to int fails
		}

		for _, tc := range complexCases {
			err := schema.Apply(map[string]interface{}{tc.field: tc.value})
			if tc.shouldPass {
				assert.NoError(t, err, "Should pass for %s: %T", tc.field, tc.value)
			} else {
				assert.Error(t, err, "Should fail for %s: %T", tc.field, tc.value)
			}
		}
	})

	t.Run("empty string to numeric", func(t *testing.T) {
		var intValue int
		var floatValue float64

		schema := NewSchema(
			Value("int", &intValue),
			Value("float", &floatValue),
		)

		// Test with empty strings - go-convert converts empty strings to 0
		err := schema.Apply(map[string]interface{}{
			"int":   "",
			"float": "",
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, intValue)
		assert.Equal(t, 0.0, floatValue)
	})

	t.Run("whitespace string to numeric", func(t *testing.T) {
		var intValue int
		var floatValue float64

		schema := NewSchema(
			Value("int", &intValue),
			Value("float", &floatValue),
		)

		whitespaceCases := []string{" ", "\t", "\n", "\r", "  \t\n\r  "}

		for _, ws := range whitespaceCases {
			err := schema.Apply(map[string]interface{}{"int": ws})
			assert.Error(t, err, "Should fail for whitespace: %q", ws)
		}
	})

	t.Run("reflect value conversion", func(t *testing.T) {
		var stringValue string
		var intValue int

		schema := NewSchema(
			Value("string", &stringValue),
			Value("int", &intValue),
		)

		reflectCases := []struct {
			field      string
			value      reflect.Value
			shouldPass bool
		}{
			{"string", reflect.ValueOf("hello"), true}, // go-convert can convert reflect.Value to string
			{"int", reflect.ValueOf(42), false},        // go-convert cannot convert reflect.Value to int
			{"string", reflect.ValueOf(123), true},     // go-convert can convert reflect.Value to string
			{"int", reflect.ValueOf("42"), false},      // go-convert cannot convert reflect.Value to int
		}

		for _, tc := range reflectCases {
			err := schema.Apply(map[string]interface{}{tc.field: tc.value})
			if tc.shouldPass {
				assert.NoError(t, err, "Should pass for %s: %v", tc.field, tc.value)
			} else {
				assert.Error(t, err, "Should fail for %s: %v", tc.field, tc.value)
			}
		}
	})

	t.Run("sql null types conversion", func(t *testing.T) {
		var nullString sql.NullString
		var nullInt sql.NullInt64
		var nullFloat sql.NullFloat64
		var nullBool sql.NullBool

		schema := NewSchema(
			Value("nullString", &nullString),
			Value("nullInt", &nullInt),
			Value("nullFloat", &nullFloat),
			Value("nullBool", &nullBool),
		)

		// Test with valid values
		err := schema.Apply(map[string]interface{}{
			"nullString": "test",
			"nullInt":    42,
			"nullFloat":  3.14,
			"nullBool":   true,
		})
		assert.NoError(t, err)
		assert.True(t, nullString.Valid)
		assert.Equal(t, "test", nullString.String)
		assert.True(t, nullInt.Valid)
		assert.Equal(t, int64(42), nullInt.Int64)
		assert.True(t, nullFloat.Valid)
		assert.Equal(t, 3.14, nullFloat.Float64)
		assert.True(t, nullBool.Valid)
		assert.Equal(t, true, nullBool.Bool)

		// Test with null values
		err = schema.Apply(map[string]interface{}{
			"nullString": nil,
			"nullInt":    nil,
			"nullFloat":  nil,
			"nullBool":   nil,
		})
		assert.NoError(t, err)
		assert.True(t, nullString.Valid) // go-convert converts nil to empty string
		assert.True(t, nullInt.Valid)    // go-convert converts nil to 0
		assert.True(t, nullFloat.Valid)  // go-convert converts nil to 0.0
		assert.True(t, nullBool.Valid)   // go-convert converts nil to false
	})

	t.Run("numeric precision edge cases", func(t *testing.T) {
		var floatValue float64
		var intValue int64

		schema := NewSchema(
			Value("float", &floatValue),
			Value("int", &intValue),
		)

		precisionCases := []struct {
			field      string
			value      interface{}
			shouldPass bool
		}{
			{"float", math.MaxFloat64, true},
			{"float", math.SmallestNonzeroFloat64, true},
			{"float", math.Inf(1), true},
			{"float", math.Inf(-1), true},
			{"float", math.NaN(), true}, // go-convert can handle NaN
			{"int", math.MaxInt64, true},
			{"int", math.MinInt64, true},
		}

		for _, tc := range precisionCases {
			err := schema.Apply(map[string]interface{}{tc.field: tc.value})
			if tc.shouldPass {
				assert.NoError(t, err, "Should pass for %s: %v", tc.field, tc.value)
			} else {
				assert.Error(t, err, "Should fail for %s: %v", tc.field, tc.value)
			}
		}
	})
}

func TestPerformance_EdgeCases(t *testing.T) {
	t.Run("large string processing", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(ToUpper(), TrimSpace())),
		)

		// Create a very large string
		largeString := strings.Repeat("a", 1000000) // 1MB string

		// Measure processing time
		start := time.Now()
		err := schema.Apply(map[string]interface{}{"test": largeString})
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, strings.ToUpper(largeString), value)
		assert.Less(t, duration, 2*time.Second, "Processing 1MB string should take less than 2 seconds")
	})

	t.Run("large slice with complex validation", func(t *testing.T) {
		type ComplexItem struct {
			ID    int
			Name  string
			Email string
			Tags  []string
		}

		var items []ComplexItem
		sliceField := Slice("items", &items)

		// Configure the callback for the slice
		if sf, ok := sliceField.(*SliceField[ComplexItem]); ok {
			sf.SetCallback(func(s *Schema, item *ComplexItem) {
				WithSchema(s, Value("id", &item.ID, WithValidators(Required(), Min(1))))
				WithSchema(s, Value("name", &item.Name, WithValidators(Required(), MinLength(1), MaxLength(100))))
				WithSchema(s, Value("email", &item.Email, WithValidators(Required(), Email())))
				WithSchema(s, Slice("tags", &item.Tags, WithValidators(MaxLength(10))))
			})
		}

		schema := NewSchema(sliceField)

		// Create a large volume of data
		largeData := make([]interface{}, 10000)
		for i := 0; i < 10000; i++ {
			largeData[i] = map[string]interface{}{
				"id":    i + 1,
				"name":  fmt.Sprintf("Item %d", i),
				"email": fmt.Sprintf("item%d@example.com", i),
				"tags":  []interface{}{"tag1", "tag2", "tag3"},
			}
		}

		// Measure processing time
		start := time.Now()
		err := schema.Apply(map[string]interface{}{"items": largeData})
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Len(t, items, 10000)
		assert.Less(t, duration, 10*time.Second, "Processing 10k complex items should take less than 10 seconds")
	})

	t.Run("deep nested structure performance", func(t *testing.T) {
		type Level5 struct {
			Value string
		}
		type Level4 struct {
			Level5 Level5
		}
		type Level3 struct {
			Level4 Level4
		}
		type Level2 struct {
			Level3 Level3
		}
		type Level1 struct {
			Level2 Level2
		}
		type DeepStruct struct {
			Level1 Level1
		}

		var deep DeepStruct
		schema := NewSchema(
			Struct("level1", &deep.Level1, WithSubSchema(func(s *Schema, level1 *Level1) {
				WithSchema(s, Struct("level2", &level1.Level2, WithSubSchema(func(s2 *Schema, level2 *Level2) {
					WithSchema(s2, Struct("level3", &level2.Level3, WithSubSchema(func(s3 *Schema, level3 *Level3) {
						WithSchema(s3, Struct("level4", &level3.Level4, WithSubSchema(func(s4 *Schema, level4 *Level4) {
							WithSchema(s4, Struct("level5", &level4.Level5, WithSubSchema(func(s5 *Schema, level5 *Level5) {
								WithSchema(s5, Value("value", &level5.Value, WithValidators(Required())))
							})))
						})))
					})))
				})))
			})),
		)

		// Create deeply nested data
		data := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"level4": map[string]interface{}{
							"level5": map[string]interface{}{
								"value": "deep value",
							},
						},
					},
				},
			},
		}

		// Measure processing time
		start := time.Now()
		err := schema.Apply(data)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, "deep value", deep.Level1.Level2.Level3.Level4.Level5.Value)
		assert.Less(t, duration, 1*time.Second, "Processing deep nested structure should be fast")
	})

	t.Run("memory usage with large maps", func(t *testing.T) {
		var configs map[string]string
		schema := NewSchema(
			Map("configs", &configs),
		)

		// Create a large volume of data
		largeData := make(map[string]interface{}, 10000)
		for i := 0; i < 10000; i++ {
			largeData[fmt.Sprintf("config_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		// Measure processing time
		start := time.Now()
		err := schema.Apply(map[string]interface{}{"configs": largeData})
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Len(t, configs, 10000)
		assert.Less(t, duration, 10*time.Second, "Processing 10k map entries should take less than 10 seconds")
	})

	t.Run("concurrent access simulation", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(ToUpper(), TrimSpace())),
		)

		// Simulate concurrent access by processing multiple values rapidly
		testData := []string{
			"hello world",
			"test string",
			"another test",
			"final test",
		}

		start := time.Now()
		for _, data := range testData {
			err := schema.Apply(map[string]interface{}{"test": data})
			assert.NoError(t, err)
		}
		duration := time.Since(start)

		assert.Less(t, duration, 1*time.Second, "Processing multiple values should be fast")
	})
}
