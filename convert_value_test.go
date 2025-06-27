package poxxy

import (
	"testing"
)

func TestConvertValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
		wantErr  bool
	}{
		// String conversions
		{
			name:     "string to string direct",
			input:    "hello",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "int to string",
			input:    42,
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "float to string",
			input:    3.14,
			expected: "3.14",
			wantErr:  false,
		},
		{
			name:     "bool to string",
			input:    true,
			expected: "true",
			wantErr:  false,
		},

		// Int conversions
		{
			name:     "int to int direct",
			input:    42,
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "int64 to int",
			input:    int64(42),
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "float64 to int",
			input:    42.0,
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "string to int",
			input:    "42",
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "invalid string to int",
			input:    "not a number",
			expected: 0,
			wantErr:  true,
		},

		// Int64 conversions
		{
			name:     "int64 to int64 direct",
			input:    int64(42),
			expected: int64(42),
			wantErr:  false,
		},
		{
			name:     "int to int64",
			input:    42,
			expected: int64(42),
			wantErr:  false,
		},
		{
			name:     "float64 to int64",
			input:    42.0,
			expected: int64(42),
			wantErr:  false,
		},
		{
			name:     "string to int64",
			input:    "42",
			expected: int64(42),
			wantErr:  false,
		},

		// Float64 conversions
		{
			name:     "float64 to float64 direct",
			input:    3.14,
			expected: 3.14,
			wantErr:  false,
		},
		{
			name:     "int to float64",
			input:    42,
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "int64 to float64",
			input:    int64(42),
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "string to float64",
			input:    "3.14",
			expected: 3.14,
			wantErr:  false,
		},
		{
			name:     "invalid string to float64",
			input:    "not a number",
			expected: 0.0,
			wantErr:  true,
		},

		// Bool conversions
		{
			name:     "bool to bool direct",
			input:    true,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string true to bool",
			input:    "true",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string TRUE to bool",
			input:    "TRUE",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string 1 to bool",
			input:    "1",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string yes to bool",
			input:    "yes",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string y to bool",
			input:    "y",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string on to bool",
			input:    "on",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string t to bool",
			input:    "t",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string false to bool",
			input:    "false",
			expected: false,
			wantErr:  false,
		},
		{
			name:     "string 0 to bool",
			input:    "0",
			expected: false,
			wantErr:  false,
		},
		{
			name:     "string no to bool",
			input:    "no",
			expected: false,
			wantErr:  false,
		},

		// Error cases
		{
			name:     "unsupported type conversion",
			input:    []int{1, 2, 3},
			expected: "[1 2 3]",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case string:
				result, err := convertValue[string](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("convertValue() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && result != expected {
					t.Errorf("convertValue() = %v, want %v", result, expected)
				}
			case int:
				result, err := convertValue[int](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("convertValue() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && result != expected {
					t.Errorf("convertValue() = %v, want %v", result, expected)
				}
			case int64:
				result, err := convertValue[int64](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("convertValue() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && result != expected {
					t.Errorf("convertValue() = %v, want %v", result, expected)
				}
			case float64:
				result, err := convertValue[float64](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("convertValue() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && result != expected {
					t.Errorf("convertValue() = %v, want %v", result, expected)
				}
			case bool:
				result, err := convertValue[bool](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("convertValue() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && result != expected {
					t.Errorf("convertValue() = %v, want %v", result, expected)
				}
			}
		})
	}
}

func TestWasAssignedMechanism_AllFieldTypes(t *testing.T) {
	t.Run("ValueField nil triggers required", func(t *testing.T) {
		var v string
		schema := NewSchema(Value("foo", &v, WithValidators(Required())))
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil || err.Error() != "foo: field is required" {
			t.Errorf("expected required error, got %v", err)
		}
	})

	t.Run("SliceField nil triggers required", func(t *testing.T) {
		var v []string
		schema := NewSchema(Slice("foo", &v, WithValidators(Required())))
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil || err.Error() != "foo: field is required" {
			t.Errorf("expected required error, got %v", err)
		}
	})

	t.Run("ArrayField nil triggers required", func(t *testing.T) {
		var v [2]string
		schema := NewSchema(Array[string]("foo", &v, WithValidators(Required())))
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil || err.Error() != "foo: field is required" {
			t.Errorf("expected required error, got %v", err)
		}
	})

	t.Run("PointerField nil triggers required", func(t *testing.T) {
		var v *string
		schema := NewSchema(Pointer("foo", &v, WithValidators(Required())))
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil || err.Error() != "foo: field is required" {
			t.Errorf("expected required error, got %v", err)
		}
	})

	t.Run("StructField nil triggers required", func(t *testing.T) {
		type S struct{ X string }
		var v S
		schema := NewSchema(Struct("foo", &v, WithValidators(Required()), WithSubSchema(func(s *Schema, v *S) {})))
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil || err.Error() != "foo: field is required" {
			t.Errorf("expected required error, got %v", err)
		}
	})

	t.Run("MapField nil triggers required", func(t *testing.T) {
		var v map[string]string
		schema := NewSchema(Map("foo", &v, WithValidators(Required())))
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil || err.Error() != "foo: field is required" {
			t.Errorf("expected required error, got %v", err)
		}
	})

	t.Run("TransformField nil triggers required", func(t *testing.T) {
		var v int
		schema := NewSchema(Transform[string, int]("foo", &v, func(s string) (int, error) { return 0, nil }, WithValidators(Required())))
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil || err.Error() != "foo: field is required" {
			t.Errorf("expected required error, got %v", err)
		}
	})

	t.Run("ValueWithoutAssignField nil triggers required", func(t *testing.T) {
		field := ValueWithoutAssign[string]("foo", WithValidators(Required()))
		schema := NewSchema(field)
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil || err.Error() != "foo: field is required" {
			t.Errorf("expected required error, got %v", err)
		}
	})

	t.Run("NestedMapField nil triggers required", func(t *testing.T) {
		var v map[string]string
		schema := NewSchema(NestedMap("foo", &v, func(s *Schema, k string, v *string) {}), Value("foo", &v, WithValidators(Required())))
		err := schema.Apply(map[string]interface{}{"foo": nil})
		if err == nil {
			t.Errorf("expected required error, got %v", err)
		}
	})
}
