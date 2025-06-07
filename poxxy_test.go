package poxxy

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
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

func TestSchema_ApplyHTTPRequest(t *testing.T) {
	tests := []struct {
		name        string
		setupSchema func() *Schema
		request     *http.Request
		wantErr     bool
		expectedErr string
	}{
		{
			name: "JSON content type with valid data",
			setupSchema: func() *Schema {
				var name string
				var age int
				return NewSchema(
					Value("name", &name),
					Value("age", &age),
				)
			},
			request: func() *http.Request {
				body := `{"name": "John", "age": 30}`
				req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			wantErr: false,
		},
		{
			name: "JSON content type with invalid JSON",
			setupSchema: func() *Schema {
				var name string
				return NewSchema(
					Value("name", &name),
				)
			},
			request: func() *http.Request {
				body := `{"name": "John", invalid json}`
				req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			wantErr:     true,
			expectedErr: "failed to unmarshal request body",
		},
		{
			name: "form-urlencoded content type with valid data",
			setupSchema: func() *Schema {
				var name string
				var age int
				return NewSchema(
					Value("name", &name),
					Value("age", &age),
				)
			},
			request: func() *http.Request {
				form := url.Values{}
				form.Add("name", "John")
				form.Add("age", "30")
				req, _ := http.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			wantErr: false,
		},
		{
			name: "unsupported content type",
			setupSchema: func() *Schema {
				var name string
				return NewSchema(
					Value("name", &name),
				)
			},
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "/test", strings.NewReader("some data"))
				req.Header.Set("Content-Type", "text/plain")
				return req
			}(),
			wantErr:     true,
			expectedErr: "unsupported content type: text/plain",
		},
		{
			name: "empty content type defaults to unsupported",
			setupSchema: func() *Schema {
				var name string
				return NewSchema(
					Value("name", &name),
				)
			},
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "/test", strings.NewReader("some data"))
				return req
			}(),
			wantErr:     true,
			expectedErr: "unsupported content type:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			err := schema.ApplyHTTPRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Schema.ApplyHTTPRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Schema.ApplyHTTPRequest() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
			}
		})
	}
}
