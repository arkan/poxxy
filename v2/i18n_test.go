package poxxy

import (
	"testing"

	"golang.org/x/text/language"
)

func TestMessageCatalog_Get(t *testing.T) {
	catalog := NewMessageCatalog(nil)

	// Test default message
	msg := catalog.Get(MsgRequired)
	if msg != "{field} is required" {
		t.Errorf("expected default message, got %q", msg)
	}

	// Test custom message
	catalog.Set(MsgRequired, "custom required message")
	msg = catalog.Get(MsgRequired)
	if msg != "custom required message" {
		t.Errorf("expected custom message, got %q", msg)
	}

	// Test unknown key
	msg = catalog.Get("unknown_key")
	if msg != "unknown_key" {
		t.Errorf("expected key as fallback, got %q", msg)
	}
}

func TestMessageCatalog_Format(t *testing.T) {
	catalog := GetDefaultCatalog()

	tests := []struct {
		name     string
		key      MessageKey
		params   map[string]any
		expected string
	}{
		{
			name:     "required",
			key:      MsgRequired,
			params:   map[string]any{"field": "email"},
			expected: "email is required",
		},
		{
			name:     "min with value",
			key:      MsgMin,
			params:   map[string]any{"field": "age", "min": 18},
			expected: "age must be at least 18",
		},
		{
			name:     "minlength",
			key:      MsgMinLength,
			params:   map[string]any{"field": "password", "min": 8},
			expected: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := catalog.Format(tt.key, tt.params)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFrenchCatalog(t *testing.T) {
	catalog := NewFrenchCatalog()

	tests := []struct {
		key      MessageKey
		params   map[string]any
		expected string
	}{
		{
			key:      MsgRequired,
			params:   map[string]any{"field": "email"},
			expected: "email est requis",
		},
		{
			key:      MsgEmail,
			params:   map[string]any{"field": "email"},
			expected: "email doit être une adresse email valide",
		},
		{
			key:      MsgMinLength,
			params:   map[string]any{"field": "mot de passe", "min": 8},
			expected: "mot de passe doit contenir au moins 8 caractères",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			result := catalog.Format(tt.key, tt.params)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSpanishCatalog(t *testing.T) {
	catalog := NewSpanishCatalog()

	msg := catalog.Format(MsgRequired, map[string]any{"field": "nombre"})
	if msg != "nombre es requerido" {
		t.Errorf("expected Spanish message, got %q", msg)
	}
}

func TestGermanCatalog(t *testing.T) {
	catalog := NewGermanCatalog()

	msg := catalog.Format(MsgRequired, map[string]any{"field": "Name"})
	if msg != "Name ist erforderlich" {
		t.Errorf("expected German message, got %q", msg)
	}
}

func TestCatalogForLanguage(t *testing.T) {
	tests := []struct {
		tag      language.Tag
		key      MessageKey
		field    string
		contains string
	}{
		{language.French, MsgRequired, "email", "est requis"},
		{language.Spanish, MsgRequired, "email", "es requerido"},
		{language.German, MsgRequired, "email", "ist erforderlich"},
		{language.English, MsgRequired, "email", "is required"},
		{language.Japanese, MsgRequired, "email", "is required"}, // Fallback to English
	}

	for _, tt := range tests {
		t.Run(tt.tag.String(), func(t *testing.T) {
			catalog := CatalogForLanguage(tt.tag)
			msg := catalog.Format(tt.key, map[string]any{"field": tt.field})
			if msg != tt.field+" "+tt.contains {
				// Check if it contains the expected phrase
				expected := tt.field + " " + tt.contains
				if msg != expected {
					t.Errorf("expected %q, got %q", expected, msg)
				}
			}
		})
	}
}

func TestMessageCatalog_Chaining(t *testing.T) {
	catalog := NewMessageCatalog(nil).
		Set(MsgRequired, "custom required").
		Set(MsgEmail, "custom email")

	if catalog.Get(MsgRequired) != "custom required" {
		t.Error("chaining Set did not work for required")
	}
	if catalog.Get(MsgEmail) != "custom email" {
		t.Error("chaining Set did not work for email")
	}
}

func TestDefaultMessages(t *testing.T) {
	// Ensure all message keys have default messages
	keys := []MessageKey{
		MsgRequired, MsgEmail, MsgURL, MsgUUID,
		MsgMin, MsgMax, MsgMinLength, MsgMaxLength,
		MsgMinItems, MsgMaxItems, MsgPattern, MsgIn,
		MsgEach, MsgUnique, MsgUniqueBy, MsgCustom,
		MsgConversion, MsgType, MsgNegative, MsgPositive,
		MsgNonZero, MsgNotEmpty,
	}

	for _, key := range keys {
		if _, ok := DefaultMessages[string(key)]; !ok {
			t.Errorf("missing default message for key %q", key)
		}
	}
}
