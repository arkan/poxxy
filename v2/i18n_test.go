package poxxy

import (
	"strings"
	"testing"

	"github.com/arkan/poxxy/v2/i18n"
	"golang.org/x/text/language"
)

// =============================================================================
// Schema Integration Tests with i18n
// =============================================================================

func TestSchemaWithFrenchTranslator(t *testing.T) {
	var email string

	schema := NewSchema(
		Field("email", &email).Validate(Required[string](), Email()),
	).WithTranslator(i18n.French())

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

func TestSchemaWithSpanishTranslator(t *testing.T) {
	var age int

	schema := NewSchema(
		Field("age", &age).Validate(Min(18)),
	).WithTranslator(i18n.Spanish())

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

func TestSchemaWithGermanTranslator(t *testing.T) {
	var name string

	schema := NewSchema(
		Field("name", &name).Validate(MinLength[string](5)),
	).WithTranslator(i18n.German())

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

func TestSchemaWithoutTranslator(t *testing.T) {
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
	translated := err.Translate(i18n.French())

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
	translated := errs.Translate(i18n.Spanish())

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

func TestCustomTranslator(t *testing.T) {
	// Create custom catalog with custom messages
	catalog := i18n.New(nil).
		Set(i18n.MsgRequired, "Le champ {field} ne peut pas être vide").
		Set(i18n.MsgEmail, "{field} n'est pas une adresse email correcte")

	var email string

	schema := NewSchema(
		Field("email", &email).Validate(Required[string]()),
	).WithTranslator(catalog)

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

func TestForLanguageIntegration(t *testing.T) {
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
			).WithTranslator(i18n.ForLanguage(lang.tag))

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
