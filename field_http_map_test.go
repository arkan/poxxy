package poxxy

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestParseFormCollection(t *testing.T) {
	tests := []struct {
		name     string
		formData url.Values
		typeName string
		want     map[string]map[string]string
	}{
		{
			name: "parses form data with multiple fields",
			formData: url.Values{
				"attachments[0][url]":            []string{"https://example.com/doc1.pdf"},
				"attachments[0][filename]":       []string{"doc1.pdf"},
				"attachments[1][url]":            []string{"https://example.com/doc2.pdf"},
				"attachments[1][filename]":       []string{"doc2.pdf"},
				"attachments[abc-def][url]":      []string{"https://example.com/doc3.pdf"},
				"attachments[abc-def][filename]": []string{"doc3.pdf"},
			},
			typeName: "attachments",
			want: map[string]map[string]string{
				"0": {
					"url":      "https://example.com/doc1.pdf",
					"filename": "doc1.pdf",
				},
				"1": {
					"url":      "https://example.com/doc2.pdf",
					"filename": "doc2.pdf",
				},
				"abc-def": {
					"url":      "https://example.com/doc3.pdf",
					"filename": "doc3.pdf",
				},
			},
		},
		{
			name:     "handles empty form data",
			formData: url.Values{},
			typeName: "attachments",
			want:     map[string]map[string]string{},
		},
		{
			name: "handles non-sequential indices",
			formData: url.Values{
				"attachments[0][url]": []string{"https://example.com/doc1.pdf"},
				"attachments[2][url]": []string{"https://example.com/doc2.pdf"},
			},
			typeName: "attachments",
			want: map[string]map[string]string{
				"0": {
					"url": "https://example.com/doc1.pdf",
				},
				"2": {
					"url": "https://example.com/doc2.pdf",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{
				PostForm: tt.formData,
			}
			got := parseFormCollection(r.PostForm, tt.typeName)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("ParseFormCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToURLValues(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		want url.Values
	}{
		{
			name: "handles string values",
			data: map[string]interface{}{
				"name": "test",
				"url":  "https://example.com",
			},
			want: url.Values{
				"name": []string{"test"},
				"url":  []string{"https://example.com"},
			},
		},
		{
			name: "handles string slice values",
			data: map[string]interface{}{
				"tags": []string{"tag1", "tag2"},
			},
			want: url.Values{
				"tags": []string{"tag1"},
			},
		},
		{
			name: "handles empty string slice",
			data: map[string]interface{}{
				"tags": []string{},
			},
			want: url.Values{
				"tags": []string{""},
			},
		},
		{
			name: "handles non-string types",
			data: map[string]interface{}{
				"count": 42,
				"price": 19.99,
				"valid": true,
			},
			want: url.Values{
				"count": []string{"42"},
				"price": []string{"19.99"},
				"valid": []string{"true"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToURLValues(tt.data)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("convertToURLValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPMap_SetDefaultValue(t *testing.T) {
	type Attachment struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
	}

	t.Run("applies default value when no form data is present", func(t *testing.T) {
		var attachments map[string]Attachment
		defaultAttachments := map[string]Attachment{
			"default": {
				URL:      "https://example.com/default.pdf",
				Filename: "default.pdf",
			},
		}

		schema := NewSchema(
			HTTPMap("attachments", &attachments, func(s *Schema, a *Attachment) {
				WithSchema(s, Value("url", &a.URL, WithValidators(Required())))
				WithSchema(s, Value("filename", &a.Filename, WithValidators(Required())))
			}, WithDefault(defaultAttachments)),
		)

		// Apply empty data - should use default value
		data := map[string]interface{}{}
		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Schema.Apply() error = %v", err)
		}

		if !reflect.DeepEqual(attachments, defaultAttachments) {
			t.Errorf("attachments = %v, want %v", attachments, defaultAttachments)
		}
	})

	t.Run("provided form data overrides default value", func(t *testing.T) {
		var attachments map[string]Attachment
		defaultAttachments := map[string]Attachment{
			"default": {
				URL:      "https://example.com/default.pdf",
				Filename: "default.pdf",
			},
		}

		schema := NewSchema(
			HTTPMap("attachments", &attachments, func(s *Schema, a *Attachment) {
				WithSchema(s, Value("url", &a.URL, WithValidators(Required())))
				WithSchema(s, Value("filename", &a.Filename, WithValidators(Required())))
			}, WithDefault(defaultAttachments)),
		)

		// Apply data with form values - should override default
		data := map[string]interface{}{
			"attachments[0][url]":      "https://example.com/doc1.pdf",
			"attachments[0][filename]": "doc1.pdf",
		}
		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Schema.Apply() error = %v", err)
		}

		expected := map[string]Attachment{
			"0": {
				URL:      "https://example.com/doc1.pdf",
				Filename: "doc1.pdf",
			},
		}

		if !reflect.DeepEqual(attachments, expected) {
			t.Errorf("attachments = %v, want %v", attachments, expected)
		}
	})

	t.Run("no default value when field is not present", func(t *testing.T) {
		var attachments map[string]Attachment

		schema := NewSchema(
			HTTPMap("attachments", &attachments, func(s *Schema, a *Attachment) {
				WithSchema(s, Value("url", &a.URL, WithValidators(Required())))
				WithSchema(s, Value("filename", &a.Filename, WithValidators(Required())))
			}),
		)

		// Apply empty data - should not set any value
		data := map[string]interface{}{}
		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Schema.Apply() error = %v", err)
		}

		if attachments != nil {
			t.Errorf("attachments = %v, want nil", attachments)
		}
	})

	t.Run("default value with complex struct", func(t *testing.T) {
		type User struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Age   int    `json:"age"`
		}

		var users map[int]User
		defaultUsers := map[int]User{
			1: {Name: "Default User", Email: "default@example.com", Age: 25},
			2: {Name: "Another User", Email: "another@example.com", Age: 30},
		}

		schema := NewSchema(
			HTTPMap("users", &users, func(s *Schema, u *User) {
				WithSchema(s, Value("name", &u.Name, WithValidators(Required())))
				WithSchema(s, Value("email", &u.Email, WithValidators(Required())))
				WithSchema(s, Value("age", &u.Age, WithValidators(Required())))
			}, WithDefault(defaultUsers)),
		)

		// Apply empty data - should use default value
		data := map[string]interface{}{}
		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Schema.Apply() error = %v", err)
		}

		if !reflect.DeepEqual(users, defaultUsers) {
			t.Errorf("users = %v, want %v", users, defaultUsers)
		}
	})

	t.Run("default value with string keys", func(t *testing.T) {
		type Config struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		}

		var configs map[string]Config
		defaultConfigs := map[string]Config{
			"theme":  {Value: "dark", Type: "string"},
			"locale": {Value: "en", Type: "string"},
		}

		schema := NewSchema(
			HTTPMap("configs", &configs, func(s *Schema, c *Config) {
				WithSchema(s, Value("value", &c.Value, WithValidators(Required())))
				WithSchema(s, Value("type", &c.Type, WithValidators(Required())))
			}, WithDefault(defaultConfigs)),
		)

		// Apply empty data - should use default value
		data := map[string]interface{}{}
		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Schema.Apply() error = %v", err)
		}

		if !reflect.DeepEqual(configs, defaultConfigs) {
			t.Errorf("configs = %v, want %v", configs, defaultConfigs)
		}
	})

	t.Run("empty default map", func(t *testing.T) {
		var attachments map[string]Attachment
		defaultAttachments := map[string]Attachment{}

		schema := NewSchema(
			HTTPMap("attachments", &attachments, func(s *Schema, a *Attachment) {
				WithSchema(s, Value("url", &a.URL, WithValidators(Required())))
				WithSchema(s, Value("filename", &a.Filename, WithValidators(Required())))
			}, WithDefault(defaultAttachments)),
		)

		// Apply empty data - should use empty default
		data := map[string]interface{}{}
		err := schema.Apply(data)
		if err != nil {
			t.Errorf("Schema.Apply() error = %v", err)
		}

		if !reflect.DeepEqual(attachments, defaultAttachments) {
			t.Errorf("attachments = %v, want %v", attachments, defaultAttachments)
		}
	})
}
