package poxxy

import (
	"strings"
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

// =============================================================================
// Schema Integration Tests
// =============================================================================

func TestSchemaWithFrenchCatalog(t *testing.T) {
	var email string

	schema := NewSchema(
		Field("email", &email).Validate(Required[string](), Email()),
	).WithCatalog(NewFrenchCatalog())

	err := schema.Parse(strings.NewReader(`{"email": ""}`))

	if err == nil {
		t.Fatal("expected validation error")
	}

	verrs, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	// Check that message is in French
	msg := verrs.Errors[0].Message
	if !strings.Contains(msg, "est requis") {
		t.Errorf("expected French message containing 'est requis', got %q", msg)
	}
}

func TestSchemaWithSpanishCatalog(t *testing.T) {
	var age int

	schema := NewSchema(
		Field("age", &age).Validate(Min(18)),
	).WithCatalog(NewSpanishCatalog())

	err := schema.Parse(strings.NewReader(`{"age": 10}`))

	if err == nil {
		t.Fatal("expected validation error")
	}

	verrs, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	// Check that message is in Spanish
	msg := verrs.Errors[0].Message
	if !strings.Contains(msg, "debe ser al menos") {
		t.Errorf("expected Spanish message containing 'debe ser al menos', got %q", msg)
	}
}

func TestSchemaWithGermanCatalog(t *testing.T) {
	var name string

	schema := NewSchema(
		Field("name", &name).Validate(MinLength[string](5)),
	).WithCatalog(NewGermanCatalog())

	err := schema.Parse(strings.NewReader(`{"name": "ab"}`))

	if err == nil {
		t.Fatal("expected validation error")
	}

	verrs, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	// Check that message is in German
	msg := verrs.Errors[0].Message
	if !strings.Contains(msg, "mindestens") {
		t.Errorf("expected German message containing 'mindestens', got %q", msg)
	}
}

func TestSchemaWithoutCatalog(t *testing.T) {
	var email string

	schema := NewSchema(
		Field("email", &email).Validate(Required[string]()),
	)

	err := schema.Parse(strings.NewReader(`{"email": ""}`))

	if err == nil {
		t.Fatal("expected validation error")
	}

	verrs, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	// Check that message is in English (default)
	msg := verrs.Errors[0].Message
	if !strings.Contains(msg, "is required") {
		t.Errorf("expected English message containing 'is required', got %q", msg)
	}
}

func TestValidationErrorTranslate(t *testing.T) {
	err := &ValidationError{
		Field:  "email",
		Value:  "",
		Rule:   "required",
		Params: map[string]any{"field": "email", "value": ""},
	}

	// Translate to French
	translated := err.Translate(NewFrenchCatalog())

	if !strings.Contains(translated.Message, "est requis") {
		t.Errorf("expected French translation, got %q", translated.Message)
	}

	// Original should be unchanged
	if err.Message == translated.Message {
		t.Error("original error should not be modified")
	}
}

func TestValidationErrorsTranslate(t *testing.T) {
	errs := &ValidationErrors{
		Errors: []*ValidationError{
			{
				Field:  "email",
				Value:  "",
				Rule:   "required",
				Params: map[string]any{"field": "email", "value": ""},
			},
			{
				Field:  "age",
				Value:  10,
				Rule:   "min",
				Params: map[string]any{"field": "age", "value": 10, "min": 18},
			},
		},
	}

	// Translate to Spanish
	translated := errs.Translate(NewSpanishCatalog())

	if len(translated.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(translated.Errors))
	}

	if !strings.Contains(translated.Errors[0].Message, "es requerido") {
		t.Errorf("expected Spanish translation for required, got %q", translated.Errors[0].Message)
	}

	if !strings.Contains(translated.Errors[1].Message, "debe ser al menos") {
		t.Errorf("expected Spanish translation for min, got %q", translated.Errors[1].Message)
	}
}

func TestCustomCatalog(t *testing.T) {
	// Create custom catalog with custom messages
	catalog := NewMessageCatalog(nil).
		Set(MsgRequired, "Le champ {field} ne peut pas être vide").
		Set(MsgEmail, "{field} n'est pas une adresse email correcte")

	var email string

	schema := NewSchema(
		Field("email", &email).Validate(Required[string]()),
	).WithCatalog(catalog)

	err := schema.Parse(strings.NewReader(`{"email": ""}`))

	if err == nil {
		t.Fatal("expected validation error")
	}

	verrs := err.(*ValidationErrors)
	msg := verrs.Errors[0].Message

	if msg != "Le champ email ne peut pas être vide" {
		t.Errorf("expected custom message, got %q", msg)
	}
}

func TestCatalogForLanguageIntegration(t *testing.T) {
	var name string

	// Test with different languages
	languages := []struct {
		tag      language.Tag
		contains string
	}{
		{language.French, "est requis"},
		{language.Spanish, "es requerido"},
		{language.German, "ist erforderlich"},
		{language.English, "is required"},
	}

	for _, lang := range languages {
		t.Run(lang.tag.String(), func(t *testing.T) {
			schema := NewSchema(
				Field("name", &name).Validate(Required[string]()),
			).WithCatalog(CatalogForLanguage(lang.tag))

			err := schema.Parse(strings.NewReader(`{"name": ""}`))
			if err == nil {
				t.Fatal("expected validation error")
			}

			verrs := err.(*ValidationErrors)
			msg := verrs.Errors[0].Message

			if !strings.Contains(msg, lang.contains) {
				t.Errorf("expected message containing %q, got %q", lang.contains, msg)
			}
		})
	}
}
