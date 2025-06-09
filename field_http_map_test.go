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
