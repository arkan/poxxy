package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestJSONDecoder(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
		wantErr  bool
	}{
		{
			name:  "simple object",
			input: `{"name": "John", "age": 30}`,
			expected: map[string]any{
				"name": "John",
				"age":  float64(30), // JSON numbers are float64
			},
		},
		{
			name:  "nested object",
			input: `{"user": {"name": "John", "email": "john@example.com"}}`,
			expected: map[string]any{
				"user": map[string]any{
					"name":  "John",
					"email": "john@example.com",
				},
			},
		},
		{
			name:  "with array",
			input: `{"tags": ["go", "validation"]}`,
			expected: map[string]any{
				"tags": []any{"go", "validation"},
			},
		},
		{
			name:     "empty object",
			input:    `{}`,
			expected: map[string]any{},
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := JSONDecoder(strings.NewReader(tt.input))

			var result map[string]any
			err := decoder.Decode(&result)

			if (err != nil) != tt.wantErr {
				t.Errorf("JSONDecoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for key, expected := range tt.expected {
					if result[key] == nil && expected != nil {
						t.Errorf("missing key %q", key)
					}
				}
			}
		})
	}
}

func TestJSONDecoderFromRequest(t *testing.T) {
	body := `{"name": "John", "age": 30}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	decoder := JSONDecoderFromRequest(req, nil)

	var result map[string]any
	err := decoder.Decode(&result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("expected name='John', got %v", result["name"])
	}
}

func TestFormDecoder(t *testing.T) {
	tests := []struct {
		name     string
		values   url.Values
		expected map[string]any
	}{
		{
			name: "simple values",
			values: url.Values{
				"name": {"John"},
				"age":  {"30"},
			},
			expected: map[string]any{
				"name": "John",
				"age":  30,
			},
		},
		{
			name: "multiple values",
			values: url.Values{
				"tags": {"go", "validation", "schema"},
			},
			expected: map[string]any{
				"tags": []any{"go", "validation", "schema"},
			},
		},
		{
			name: "boolean values",
			values: url.Values{
				"active":   {"true"},
				"disabled": {"false"},
			},
			expected: map[string]any{
				"active":   true,
				"disabled": false,
			},
		},
		{
			name: "float values",
			values: url.Values{
				"price": {"19.99"},
			},
			expected: map[string]any{
				"price": 19.99,
			},
		},
		{
			name:     "empty values",
			values:   url.Values{},
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := FormDecoder(tt.values)

			var result map[string]any
			err := decoder.Decode(&result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for key, expected := range tt.expected {
				got := result[key]
				switch exp := expected.(type) {
				case int:
					if got != exp {
						t.Errorf("key %q: expected %v, got %v", key, exp, got)
					}
				case float64:
					if got != exp {
						t.Errorf("key %q: expected %v, got %v", key, exp, got)
					}
				case bool:
					if got != exp {
						t.Errorf("key %q: expected %v, got %v", key, exp, got)
					}
				case string:
					if got != exp {
						t.Errorf("key %q: expected %v, got %v", key, exp, got)
					}
				case []any:
					gotSlice, ok := got.([]any)
					if !ok {
						t.Errorf("key %q: expected slice, got %T", key, got)
						continue
					}
					if len(gotSlice) != len(exp) {
						t.Errorf("key %q: expected %d items, got %d", key, len(exp), len(gotSlice))
					}
				}
			}
		})
	}
}

func TestFormDecoderFromRequest(t *testing.T) {
	form := url.Values{
		"name": {"John"},
		"age":  {"30"},
	}
	body := form.Encode()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	decoder, err := FormDecoderFromRequest(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	err = decoder.Decode(&result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("expected name='John', got %v", result["name"])
	}
	if result["age"] != 30 {
		t.Errorf("expected age=30, got %v", result["age"])
	}
}

func TestQueryDecoder(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected map[string]any
	}{
		{
			name:  "simple query",
			query: "name=John&age=30",
			expected: map[string]any{
				"name": "John",
				"age":  30,
			},
		},
		{
			name:  "multiple values",
			query: "tags=go&tags=validation",
			expected: map[string]any{
				"tags": []any{"go", "validation"},
			},
		},
		{
			name:  "encoded values",
			query: "message=hello%20world",
			expected: map[string]any{
				"message": "hello world",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			decoder := QueryDecoder(values)

			var result map[string]any
			err := decoder.Decode(&result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for key, expected := range tt.expected {
				got := result[key]
				switch exp := expected.(type) {
				case int:
					if got != exp {
						t.Errorf("key %q: expected %v, got %v", key, exp, got)
					}
				case string:
					if got != exp {
						t.Errorf("key %q: expected %v, got %v", key, exp, got)
					}
				}
			}
		})
	}
}

func TestQueryDecoderFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?name=John&age=30", nil)

	decoder := QueryDecoderFromRequest(req)

	var result map[string]any
	err := decoder.Decode(&result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("expected name='John', got %v", result["name"])
	}
	if result["age"] != 30 {
		t.Errorf("expected age=30, got %v", result["age"])
	}
}

func TestDecoderFromRequest(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		query       string
		wantErr     bool
	}{
		{
			name:        "JSON POST",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"name": "John"}`,
		},
		{
			name:        "JSON with charset",
			method:      http.MethodPost,
			contentType: "application/json; charset=utf-8",
			body:        `{"name": "John"}`,
		},
		{
			name:        "Form POST",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			body:        "name=John",
		},
		{
			name:   "GET with query",
			method: http.MethodGet,
			query:  "name=John",
		},
		{
			name:        "unsupported content type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        "hello",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			url := "/"
			if tt.query != "" {
				url += "?" + tt.query
			}

			req := httptest.NewRequest(tt.method, url, body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			decoder, err := DecoderFromRequest(req, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecoderFromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && decoder == nil {
				t.Error("expected decoder, got nil")
			}
		})
	}
}

func TestDecoderFromRequestWithOptions(t *testing.T) {
	body := `{"name": "John"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Wrong content type

	// Force JSON decoding
	opts := &Options{
		ContentType: ContentTypeJSON,
		MaxBodySize: MaxBodySize,
	}

	decoder, err := DecoderFromRequest(req, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	err = decoder.Decode(&result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("expected name='John', got %v", result["name"])
	}
}

func TestNestedFormDecoder(t *testing.T) {
	tests := []struct {
		name     string
		values   url.Values
		expected map[string]any
	}{
		{
			name: "simple nested",
			values: url.Values{
				"user[name]":  {"John"},
				"user[email]": {"john@example.com"},
			},
			expected: map[string]any{
				"user": map[string]any{
					"name":  "John",
					"email": "john@example.com",
				},
			},
		},
		{
			name: "deep nested",
			values: url.Values{
				"user[address][street]": {"123 Main St"},
				"user[address][city]":   {"Springfield"},
			},
			expected: map[string]any{
				"user": map[string]any{
					"address": map[string]any{
						"street": "123 Main St",
						"city":   "Springfield",
					},
				},
			},
		},
		{
			name: "array-like keys",
			values: url.Values{
				"items[0][name]": {"Item 1"},
				"items[1][name]": {"Item 2"},
			},
			expected: map[string]any{
				"items": map[string]any{
					"0": map[string]any{"name": "Item 1"},
					"1": map[string]any{"name": "Item 2"},
				},
			},
		},
		{
			name: "mixed flat and nested",
			values: url.Values{
				"name":          {"John"},
				"address[city]": {"Springfield"},
			},
			expected: map[string]any{
				"name": "John",
				"address": map[string]any{
					"city": "Springfield",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NestedFormDecoder(tt.values)

			var result map[string]any
			err := decoder.Decode(&result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Deep check
			checkNestedMap(t, "", tt.expected, result)
		})
	}
}

func checkNestedMap(t *testing.T, path string, expected, got map[string]any) {
	t.Helper()

	for key, expVal := range expected {
		fullPath := key
		if path != "" {
			fullPath = path + "." + key
		}

		gotVal, ok := got[key]
		if !ok {
			t.Errorf("missing key at %s", fullPath)
			continue
		}

		switch exp := expVal.(type) {
		case map[string]any:
			gotMap, ok := gotVal.(map[string]any)
			if !ok {
				t.Errorf("at %s: expected map, got %T", fullPath, gotVal)
				continue
			}
			checkNestedMap(t, fullPath, exp, gotMap)
		case string:
			if gotVal != exp {
				t.Errorf("at %s: expected %q, got %q", fullPath, exp, gotVal)
			}
		default:
			if gotVal != expVal {
				t.Errorf("at %s: expected %v, got %v", fullPath, expVal, gotVal)
			}
		}
	}
}

func TestParseNestedKey(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"name", []string{"name"}},
		{"user[name]", []string{"user", "name"}},
		{"user[address][city]", []string{"user", "address", "city"}},
		{"items[0][name]", []string{"items", "0", "name"}},
		{"simple", []string{"simple"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseNestedKey(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseNestedKey(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parseNestedKey(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestConvertFormValue(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"hello", "hello"},
		{"", ""},
		{"true", true},
		{"false", false},
		{"42", 42},
		{"3.14", 3.14},
		{"-10", -10},
		{"not a number", "not a number"},
		{"123abc", "123abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertFormValue(tt.input)
			if result != tt.expected {
				t.Errorf("convertFormValue(%q) = %v (%T), want %v (%T)",
					tt.input, result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestMaxBodySize(t *testing.T) {
	// Create a large body
	largeBody := bytes.Repeat([]byte("a"), 100)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")

	opts := &Options{
		MaxBodySize: 10, // Only allow 10 bytes
		ContentType: ContentTypeJSON,
	}

	decoder := JSONDecoderFromRequest(req, opts)

	var result map[string]any
	err := decoder.Decode(&result)

	// Should fail because body is truncated
	if err == nil {
		t.Error("expected error due to truncated body")
	}
}
