package i18n

import (
	"testing"

	"golang.org/x/text/language"
)

func TestCatalog_Get(t *testing.T) {
	catalog := New(nil)

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

func TestCatalog_Format(t *testing.T) {
	catalog := Default()

	tests := []struct {
		name     string
		rule     string
		params   map[string]any
		expected string
	}{
		{
			name:     "required",
			rule:     "required",
			params:   map[string]any{"field": "email"},
			expected: "email is required",
		},
		{
			name:     "min with value",
			rule:     "min",
			params:   map[string]any{"field": "age", "min": 18},
			expected: "age must be at least 18",
		},
		{
			name:     "minlength",
			rule:     "min_length",
			params:   map[string]any{"field": "password", "min": 8},
			expected: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := catalog.Format(tt.rule, tt.params)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFrenchCatalog(t *testing.T) {
	catalog := French()

	tests := []struct {
		rule     string
		params   map[string]any
		expected string
	}{
		{
			rule:     "required",
			params:   map[string]any{"field": "email"},
			expected: "email est requis",
		},
		{
			rule:     "email",
			params:   map[string]any{"field": "email"},
			expected: "email doit être une adresse email valide",
		},
		{
			rule:     "min_length",
			params:   map[string]any{"field": "mot de passe", "min": 8},
			expected: "mot de passe doit contenir au moins 8 caractères",
		},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			result := catalog.Format(tt.rule, tt.params)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSpanishCatalog(t *testing.T) {
	catalog := Spanish()

	msg := catalog.Format("required", map[string]any{"field": "nombre"})
	if msg != "nombre es requerido" {
		t.Errorf("expected Spanish message, got %q", msg)
	}
}

func TestGermanCatalog(t *testing.T) {
	catalog := German()

	msg := catalog.Format("required", map[string]any{"field": "Name"})
	if msg != "Name ist erforderlich" {
		t.Errorf("expected German message, got %q", msg)
	}
}

func TestForLanguage(t *testing.T) {
	tests := []struct {
		tag      language.Tag
		rule     string
		field    string
		contains string
	}{
		{language.French, "required", "email", "est requis"},
		{language.Spanish, "required", "email", "es requerido"},
		{language.German, "required", "email", "ist erforderlich"},
		{language.English, "required", "email", "is required"},
		{language.Japanese, "required", "email", "is required"}, // Fallback to English
	}

	for _, tt := range tests {
		t.Run(tt.tag.String(), func(t *testing.T) {
			catalog := ForLanguage(tt.tag)
			msg := catalog.Format(tt.rule, map[string]any{"field": tt.field})
			expected := tt.field + " " + tt.contains
			if msg != expected {
				t.Errorf("expected %q, got %q", expected, msg)
			}
		})
	}
}

func TestCatalog_Chaining(t *testing.T) {
	catalog := New(nil).
		Set(MsgRequired, "custom required").
		Set(MsgEmail, "custom email")

	if catalog.Get(MsgRequired) != "custom required" {
		t.Error("chaining Set did not work for required")
	}
	if catalog.Get(MsgEmail) != "custom email" {
		t.Error("chaining Set did not work for email")
	}
}

func TestDefaultMessages_Complete(t *testing.T) {
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

func TestTranslatorInterface(t *testing.T) {
	// Verify that Catalog implements Translator interface
	var _ Translator = (*Catalog)(nil)

	catalog := French()
	msg := catalog.Format("required", map[string]any{"field": "test"})
	if msg != "test est requis" {
		t.Errorf("unexpected message: %q", msg)
	}
}
