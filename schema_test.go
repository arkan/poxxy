package poxxy

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSchema_ApplyHTTPRequest(t *testing.T) {
	tests := []struct {
		name         string
		setupSchema  func(name *string, age *int) *Schema
		request      *http.Request
		wantErr      bool
		expectedErr  string
		expectedName string
		expectedAge  int
	}{
		{
			name: "JSON content type with valid data",
			setupSchema: func(name *string, age *int) *Schema {
				return NewSchema(
					Value("name", name),
					Value("age", age),
				)
			},
			request: func() *http.Request {
				body := `{"name": "John", "age": 30}`
				req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			wantErr:      false,
			expectedName: "John",
			expectedAge:  30,
		},
		{
			name: "JSON content type with invalid JSON",
			setupSchema: func(name *string, age *int) *Schema {
				return NewSchema(
					Value("name", name),
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
			setupSchema: func(name *string, age *int) *Schema {
				return NewSchema(
					Value("name", name),
					Value("age", age),
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
			wantErr:      false,
			expectedName: "John",
			expectedAge:  30,
		},
		{
			name: "no content type",
			setupSchema: func(name *string, age *int) *Schema {
				return NewSchema(
					Value("name", name),
					Value("age", age),
				)
			},
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "/test?name=John&age=30", strings.NewReader("some data"))
				return req
			}(),
			wantErr:      false,
			expectedName: "John",
			expectedAge:  30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var name string
			var age int
			schema := tt.setupSchema(&name, &age)
			err := schema.ApplyHTTPRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Schema.ApplyHTTPRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Schema.ApplyHTTPRequest() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
			}
			if !tt.wantErr {
				if name != tt.expectedName {
					t.Errorf("Schema.ApplyHTTPRequest() name = %v, want %v", name, tt.expectedName)
				}
				if age != tt.expectedAge {
					t.Errorf("Schema.ApplyHTTPRequest() age = %v, want %v", age, tt.expectedAge)
				}
			}
		})
	}
}

func TestSchema_ApplyJSON(t *testing.T) {
	tests := []struct {
		name        string
		setupSchema func() *Schema
		jsonData    string
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid json",
			setupSchema: func() *Schema {
				var name string
				return NewSchema(
					Value("name", &name),
				)
			},
			jsonData: `{"name": "test"}`,
			wantErr:  false,
		},
		{
			name: "invalid json",
			setupSchema: func() *Schema {
				var name string
				return NewSchema(
					Value("name", &name),
				)
			},
			jsonData:    `{"name": "test"`,
			wantErr:     true,
			expectedErr: "failed to unmarshal request body",
		},
		{
			name: "validation error",
			setupSchema: func() *Schema {
				var age int
				return NewSchema(
					Value("age", &age, WithValidators(Required())),
				)
			},
			jsonData:    `{"name": "test"}`,
			wantErr:     true,
			expectedErr: "age: field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			err := schema.ApplyJSON([]byte(tt.jsonData))
			if (err != nil) != tt.wantErr {
				t.Errorf("Schema.ApplyJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Schema.ApplyJSON() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
			}
		})
	}
}
