package i18n

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Translator is the interface that catalogs must implement for use with Schema.
type Translator interface {
	// Format returns a formatted message for the given rule and parameters.
	Format(rule string, params map[string]any) string
}

// MessagePrinter defines the interface for message printing with translations.
type MessagePrinter interface {
	Sprintf(key message.Reference, args ...any) string
}

// MessageKey is a key for looking up translated messages.
type MessageKey string

// Common message keys for built-in validators.
const (
	MsgRequired   MessageKey = "required"
	MsgNotEmpty   MessageKey = "not_empty"
	MsgEmail      MessageKey = "email"
	MsgURL        MessageKey = "url"
	MsgUUID       MessageKey = "uuid"
	MsgMin        MessageKey = "min"
	MsgMax        MessageKey = "max"
	MsgMinLength  MessageKey = "min_length"
	MsgMaxLength  MessageKey = "max_length"
	MsgMinItems   MessageKey = "min_items"
	MsgMaxItems   MessageKey = "max_items"
	MsgPattern    MessageKey = "pattern"
	MsgIn         MessageKey = "in"
	MsgNotIn      MessageKey = "not_in"
	MsgEach       MessageKey = "each"
	MsgUnique     MessageKey = "unique"
	MsgUniqueBy   MessageKey = "unique_by"
	MsgCustom     MessageKey = "custom"
	MsgDeferred   MessageKey = "deferred"
	MsgConversion MessageKey = "conversion"
	MsgType       MessageKey = "type"
	MsgTransform  MessageKey = "transform"
	MsgConvert    MessageKey = "convert"
	MsgRequiredIf MessageKey = "required_if"
	MsgNegative   MessageKey = "negative"
	MsgPositive   MessageKey = "positive"
	MsgNonZero    MessageKey = "non_zero"
)

// Catalog holds translations for validation messages.
type Catalog struct {
	printer  MessagePrinter
	messages map[string]string
}

// New creates a new message catalog with the given printer.
func New(printer MessagePrinter) *Catalog {
	return &Catalog{
		printer:  printer,
		messages: make(map[string]string),
	}
}

// Set sets a message for a key.
func (c *Catalog) Set(key MessageKey, message string) *Catalog {
	c.messages[string(key)] = message
	return c
}

// Get returns the message for a key, falling back to DefaultMessages.
func (c *Catalog) Get(key MessageKey) string {
	if msg, ok := c.messages[string(key)]; ok {
		return msg
	}
	if msg, ok := DefaultMessages[string(key)]; ok {
		return msg
	}
	return string(key)
}

// Format formats a message with the given parameters.
// Implements the Translator interface.
func (c *Catalog) Format(rule string, params map[string]any) string {
	template := c.Get(MessageKey(rule))
	return interpolateTemplate(template, params)
}

// interpolateTemplate replaces placeholders in a message template.
func interpolateTemplate(template string, params map[string]any) string {
	if template == "" {
		return ""
	}

	result := template
	for key, val := range params {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
	}

	return result
}

// defaultPrinter is the default English printer.
var defaultPrinter = message.NewPrinter(language.English)

// defaultCatalog is the default message catalog using English.
var defaultCatalog = &Catalog{
	printer:  defaultPrinter,
	messages: DefaultMessages,
}

// Default returns the default English message catalog.
func Default() *Catalog {
	return defaultCatalog
}

// French creates a message catalog with French translations.
func French() *Catalog {
	return &Catalog{
		printer:  message.NewPrinter(language.French),
		messages: FrenchMessages,
	}
}

// Spanish creates a message catalog with Spanish translations.
func Spanish() *Catalog {
	return &Catalog{
		printer:  message.NewPrinter(language.Spanish),
		messages: SpanishMessages,
	}
}

// German creates a message catalog with German translations.
func German() *Catalog {
	return &Catalog{
		printer:  message.NewPrinter(language.German),
		messages: GermanMessages,
	}
}

// ForLanguage returns a message catalog for the given language tag.
// Falls back to English if the language is not supported.
func ForLanguage(tag language.Tag) *Catalog {
	base, _ := tag.Base()
	switch base.String() {
	case "fr":
		return French()
	case "es":
		return Spanish()
	case "de":
		return German()
	default:
		return defaultCatalog
	}
}
