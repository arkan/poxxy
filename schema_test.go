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
	"github.com/stretchr/testify/require"
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

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
				assert.Equal(t, tt.expectedAge, age)
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

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
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

	assert.Error(t, err)
	errs, ok := err.(Errors)
	assert.True(t, ok)
	assert.Len(t, errs, 1)
	assert.Equal(t, "name description", errs[0].Description)
}

func TestSchema_ApplyWithConvert(t *testing.T) {

	t.Run("convert with required validator", func(t *testing.T) {
		var timestamp time.Time
		var normalizedEmail string

		schema := NewSchema(
			// Transform Unix timestamp to time.Time
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				t := time.Unix(unixTime, 0)
				return &t, nil
			}, WithValidators(Required())),

			// Normalize email to lowercase
			Convert[string, string]("email", &normalizedEmail, func(email string) (*string, error) {
				s := strings.ToLower(strings.TrimSpace(email))
				return &s, nil
			}, WithValidators(Required(), Email())),
		)

		data := map[string]interface{}{
			// "created_at": 1717689600, // We skip it.
			"email": "John.Doe@example.com",
		}

		err := schema.Apply(data)
		assert.Error(t, err)
		errs, ok := err.(Errors)
		assert.True(t, ok)
		assert.Len(t, errs, 1)
		assert.Equal(t, "field is required", errs[0].Error.Error())
	})

	t.Run("convert with default value", func(t *testing.T) {
		var timestamp time.Time
		var normalizedEmail string

		schema := NewSchema(
			// Transform Unix timestamp to time.Time
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				t := time.Unix(unixTime, 0)
				return &t, nil
			}, WithValidators(Required()), WithDefault(time.Unix(1717689600, 0))),

			// Normalize email to lowercase
			Convert[string, string]("email", &normalizedEmail, func(email string) (*string, error) {
				s := strings.ToLower(strings.TrimSpace(email))
				return &s, nil
			}, WithValidators(Required(), Email())),
		)

		data := map[string]interface{}{
			// "created_at": 1717689600, // We skip it.
			"email": "John.Doe@example.com",
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, int64(1717689600), timestamp.Unix())
		assert.Equal(t, "john.doe@example.com", normalizedEmail)
	})

	t.Run("convert with default value on a pointer", func(t *testing.T) {

		var timestamp *time.Time
		var normalizedEmail *string

		schema := NewSchema(
			// Transform Unix timestamp to time.Time
			ConvertPointer[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				t := time.Unix(unixTime, 0)
				return &t, nil
			}, WithValidators(Required()), WithDefault(time.Unix(1717689600, 0))),

			// Normalize email to lowercase
			ConvertPointer[string, string]("email", &normalizedEmail, func(email string) (*string, error) {
				s := strings.ToLower(strings.TrimSpace(email))
				return &s, nil
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
		assert.Error(t, err)
		errs, ok := err.(Errors)
		assert.True(t, ok)
		assert.Len(t, errs, 2)
		assert.Equal(t, "field is required", errs[0].Error.Error())
		assert.Equal(t, "field is required", errs[1].Error.Error())
	})

	t.Run("transform with custom validator", func(t *testing.T) {
		var timestamp time.Time
		var normalizedEmail string

		unix := int64(1717689600)
		schema := NewSchema(
			// Transform Unix timestamp to time.Time
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				t := time.Unix(unixTime, 0)
				return &t, nil
			}, WithValidators(Required(), ValidatorFunc(func(value time.Time, fieldName string) error {
				return fmt.Errorf("must be greater than %d", unix)
			}))),

			// Normalize email to lowercase
			Convert[string, string]("email", &normalizedEmail, func(email string) (*string, error) {
				s := strings.ToLower(strings.TrimSpace(email))
				return &s, nil
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
		assert.Len(t, errs, 1)
		assert.Equal(t, "must be greater than 1717689600", errs[0].Error.Error())
	})
}

func TestSchema_ApplyWithMultipleErrors(t *testing.T) {
	values := url.Values{
		"page":          []string{"-2"},       // 1 is the minimum page
		"limit":         []string{"99"},       // 100 is the minimum limit
		"abtdated":      []string{"20240101"}, // invalid date format
		"abtdatef":      []string{"20241231"}, // invalid date format
		"abtedi":        []string{"INVALID"},  // invalid editeur code
		"abtcdee":       []string{"INVALID"},  // invalid cdee code
		"abtreg":        []string{"INVALID"},  // invalid registration code
		"email_address": []string{"INVALID"},  // invalid email address
		"periode_type":  []string{"INVALID"},  // invalid periode type
	}

	var page int
	var limit int
	var emailAddress *string
	var periodeType *string
	var byAbtDated *time.Time
	var byAbtDatef *time.Time
	var byAbtEditeur *int64
	var byAbtreg *bool

	timeParsing := func(s string) (*time.Time, error) {
		if s == "" {
			return nil, nil
		}

		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, err
		}

		return &t, nil
	}

	schema := NewSchema(
		Value("page", &page, WithValidators(Required(), Min(1))),
		Value("limit", &limit, WithValidators(Required(), Min(100))),
		Pointer("email_address", &emailAddress, WithValidators(Required(), Email())),
		Pointer("periode_type", &periodeType, WithValidators(Required(), In("monthly", "yearly"))),
		ConvertPointer[string, time.Time]("abtdated",
			&byAbtDated,
			timeParsing,
			WithDescription("Date de début"),
			WithDefault(time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.UTC)),
		),
		ConvertPointer[string, time.Time]("abtdatef",
			&byAbtDatef,
			timeParsing,
			WithDescription("Date de fin"),
			WithDefault(time.Date(time.Now().Year(), time.December, 31, 0, 0, 0, 0, time.UTC)),
		),
		Pointer("abtedi",
			&byAbtEditeur,
			WithDescription("Code éditeur"),
		),
		Pointer("abtreg",
			&byAbtreg,
			WithDescription("Paiement"),
		),
	)

	r, err := http.NewRequest("GET", "/?"+values.Encode(), nil)
	require.NoError(t, err)
	err = schema.ApplyHTTPRequest(r)
	require.Error(t, err)
	errs, ok := err.(Errors)
	require.True(t, ok)
	require.Len(t, errs, 8)
}
