package poxxy

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestSchema_ApplyWithDescription(t *testing.T) {
	var name string
	schema := NewSchema(
		Value("name", &name, WithValidators(Required(), MinLength(10)), WithDescription("name description")),
	)
	err := schema.ApplyJSON([]byte(`{"name": "test"}`))
	if err == nil {
		t.Errorf("Schema.ApplyJSON() expected an error, but got %v", err)
	}
	errs, ok := err.(Errors)
	if !ok {
		t.Errorf("Schema.ApplyJSON() expected an Errors, but got %v", err)
	}
	if len(errs) != 1 {
		t.Errorf("Schema.ApplyJSON() expected 1 error, but got %v", len(errs))
	}
	if errs[0].Description != "name description" {
		t.Errorf("Schema.ApplyJSON() expected error description to be %v, but got %v", "name description", errs[0].Description)
	}
}

func TestSchema_ApplyWithConvert(t *testing.T) {

	t.Run("convert with required validator", func(t *testing.T) {
		var timestamp time.Time
		var normalizedEmail string

		schema := NewSchema(
			// Transform Unix timestamp to time.Time
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (time.Time, error) {
				return time.Unix(unixTime, 0), nil
			}, WithValidators(Required())),

			// Normalize email to lowercase
			Convert[string, string]("email", &normalizedEmail, func(email string) (string, error) {
				return strings.ToLower(strings.TrimSpace(email)), nil
			}, WithValidators(Required(), Email())),
		)

		data := map[string]interface{}{
			// "created_at": 1717689600, // We skip it.
			"email": "John.Doe@example.com",
		}

		err := schema.Apply(data)
		if err == nil {
			t.Errorf("Schema.Apply() expected an error, but got %v", err)
		}
		errs, ok := err.(Errors)
		if !ok {
			t.Errorf("Schema.Apply() expected an Errors, but got %v", err)
		}
		if len(errs) != 1 {
			t.Errorf("Schema.Apply() expected 1 error, but got %v", len(errs))
		}
		if errs[0].Error.Error() != "field is required" {
			t.Errorf("Schema.Apply() expected error to be %v, but got %v", "field is required", errs[0].Error.Error())
		}
	})

	t.Run("convert with default value", func(t *testing.T) {
		var timestamp time.Time
		var normalizedEmail string

		schema := NewSchema(
			// Transform Unix timestamp to time.Time
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (time.Time, error) {
				return time.Unix(unixTime, 0), nil
			}, WithValidators(Required()), WithDefault(time.Unix(1717689600, 0))),

			// Normalize email to lowercase
			Convert[string, string]("email", &normalizedEmail, func(email string) (string, error) {
				return strings.ToLower(strings.TrimSpace(email)), nil
			}, WithValidators(Required(), Email())),
		)

		data := map[string]interface{}{
			// "created_at": 1717689600, // We skip it.
			"email": "John.Doe@example.com",
		}

		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Schema.Apply() expected no errors, but got %v", err)
		} else {
			if timestamp.Unix() != 1717689600 {
				t.Errorf("Schema.Apply() expected timestamp to be %v, but got %v", 1717689600, timestamp.Unix())
			}
			if normalizedEmail != "john.doe@example.com" {
				t.Errorf("Schema.Apply() expected normalizedEmail to be %v, but got %v", "john.doe@example.com", normalizedEmail)
			}
		}
	})

	t.Run("convert with default value on a pointer", func(t *testing.T) {

		var timestamp *time.Time
		var normalizedEmail *string

		schema := NewSchema(
			// Transform Unix timestamp to time.Time
			ConvertPointer[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (time.Time, error) {
				return time.Unix(unixTime, 0), nil
			}, WithValidators(Required()), WithDefault(time.Unix(1717689600, 0))),

			// Normalize email to lowercase
			ConvertPointer[string, string]("email", &normalizedEmail, func(email string) (string, error) {
				return strings.ToLower(strings.TrimSpace(email)), nil
			}, WithValidators(Required(), Email())),
		)

		data := map[string]interface{}{
			// "created_at": 1717689600, // We skip it.
			"email": "John.Doe@example.com",
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, int64(1717689600), timestamp.Unix())
		assert.Equal(t, "john.doe@example.com", *normalizedEmail)

	})

	t.Run("apply with nil values", func(t *testing.T) {
		var name string
		var age int
		schema := NewSchema(
			Value("name", &name, WithValidators(Required())),
			Value("age", &age, WithValidators(Required())),
		)
		err := schema.Apply(map[string]interface{}{"name": nil, "age": nil})
		if err == nil {
			t.Errorf("Schema.Apply() expected an error, but got %v", err)
		}
		errs, ok := err.(Errors)
		assert.True(t, ok)
		assert.Equal(t, 2, len(errs))
		assert.Equal(t, "field is required", errs[0].Error.Error())
		assert.Equal(t, "field is required", errs[1].Error.Error())
	})

	t.Run("transform with custom validator", func(t *testing.T) {
		var timestamp time.Time
		var normalizedEmail string

		unix := int64(1717689600)
		schema := NewSchema(
			// Transform Unix timestamp to time.Time
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (time.Time, error) {
				return time.Unix(unixTime, 0), nil
			}, WithValidators(Required(), ValidatorFunc(func(value time.Time, fieldName string) error {
				return fmt.Errorf("must be greater than %d", unix)
			}))),

			// Normalize email to lowercase
			Convert[string, string]("email", &normalizedEmail, func(email string) (string, error) {
				return strings.ToLower(strings.TrimSpace(email)), nil
			}, WithValidators(Required(), Email())),
		)

		data := map[string]interface{}{
			"created_at": unix, // We skip it.
			"email":      "John.Doe@example.com",
		}

		err := schema.Apply(data)
		assert.Error(t, err)
		errs, ok := err.(Errors)
		assert.True(t, ok)
		assert.Equal(t, 1, len(errs))
		assert.Equal(t, "must be greater than 1717689600", errs[0].Error.Error())
	})
}
