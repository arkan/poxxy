package poxxy

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// =============================================================================
// Internationalization Support
// =============================================================================

// MessagePrinter defines the interface for message printing with translations.
type MessagePrinter interface {
	Sprintf(key message.Reference, args ...any) string
}

// defaultPrinter is the default English printer.
var defaultPrinter = message.NewPrinter(language.English)

// DefaultMessages contains the default English messages for built-in validators.
var DefaultMessages = map[string]string{
	"required":    "{field} is required",
	"not_empty":   "{field} must not be empty",
	"email":       "{field} must be a valid email address",
	"url":         "{field} must be a valid URL",
	"uuid":        "{field} must be a valid UUID",
	"min":         "{field} must be at least {min}",
	"max":         "{field} must be at most {max}",
	"min_length":  "{field} must be at least {min} characters",
	"max_length":  "{field} must be at most {max} characters",
	"min_items":   "{field} must have at least {min} items",
	"max_items":   "{field} must have at most {max} items",
	"pattern":     "{field} does not match the required pattern",
	"in":          "{field} must be one of the allowed values",
	"not_in":      "{field} must not be one of the disallowed values",
	"each":        "{field}[{index}]: {error}",
	"unique":      "{field} contains duplicate values",
	"unique_by":   "{field} contains duplicate values for the key",
	"custom":      "{field} is invalid",
	"deferred":    "{field} validation failed",
	"conversion":  "{field}: cannot convert value",
	"type":        "{field}: expected type {expected}, got {actual}",
	"transform":   "{field}: transform error",
	"convert":     "{field}: conversion error",
	"required_if": "{field} is required",
	"negative":    "{field} must be negative",
	"positive":    "{field} must be positive",
	"non_zero":    "{field} must not be zero",
}

// =============================================================================
// Message Key Type
// =============================================================================

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

// =============================================================================
// Message Catalog
// =============================================================================

// MessageCatalog holds translations for validation messages.
type MessageCatalog struct {
	printer  MessagePrinter
	messages map[string]string
}

// NewMessageCatalog creates a new message catalog with the given printer.
func NewMessageCatalog(printer MessagePrinter) *MessageCatalog {
	return &MessageCatalog{
		printer:  printer,
		messages: make(map[string]string),
	}
}

// Set sets a message for a key.
func (c *MessageCatalog) Set(key MessageKey, message string) *MessageCatalog {
	c.messages[string(key)] = message
	return c
}

// Get returns the message for a key, falling back to DefaultMessages.
func (c *MessageCatalog) Get(key MessageKey) string {
	if msg, ok := c.messages[string(key)]; ok {
		return msg
	}
	if msg, ok := DefaultMessages[string(key)]; ok {
		return msg
	}
	return string(key)
}

// Format formats a message with the given parameters.
func (c *MessageCatalog) Format(key MessageKey, params map[string]any) string {
	template := c.Get(key)
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

// =============================================================================
// Default Catalog
// =============================================================================

// defaultCatalog is the default message catalog using English.
var defaultCatalog = &MessageCatalog{
	printer:  defaultPrinter,
	messages: DefaultMessages,
}

// GetDefaultCatalog returns the default message catalog.
func GetDefaultCatalog() *MessageCatalog {
	return defaultCatalog
}

// =============================================================================
// Language-Specific Catalogs
// =============================================================================

// FrenchMessages contains French translations.
var FrenchMessages = map[string]string{
	"required":    "{field} est requis",
	"not_empty":   "{field} ne doit pas être vide",
	"email":       "{field} doit être une adresse email valide",
	"url":         "{field} doit être une URL valide",
	"uuid":        "{field} doit être un UUID valide",
	"min":         "{field} doit être au moins {min}",
	"max":         "{field} doit être au plus {max}",
	"min_length":  "{field} doit contenir au moins {min} caractères",
	"max_length":  "{field} doit contenir au plus {max} caractères",
	"min_items":   "{field} doit contenir au moins {min} éléments",
	"max_items":   "{field} doit contenir au plus {max} éléments",
	"pattern":     "{field} ne correspond pas au format requis",
	"in":          "{field} doit être l'une des valeurs autorisées",
	"not_in":      "{field} ne doit pas être l'une des valeurs interdites",
	"each":        "{field}[{index}]: {error}",
	"unique":      "{field} contient des doublons",
	"unique_by":   "{field} contient des doublons pour la clé",
	"custom":      "{field} est invalide",
	"deferred":    "{field} validation échouée",
	"conversion":  "{field}: impossible de convertir la valeur",
	"type":        "{field}: type attendu {expected}, reçu {actual}",
	"transform":   "{field}: erreur de transformation",
	"convert":     "{field}: erreur de conversion",
	"required_if": "{field} est requis",
	"negative":    "{field} doit être négatif",
	"positive":    "{field} doit être positif",
	"non_zero":    "{field} ne doit pas être zéro",
}

// SpanishMessages contains Spanish translations.
var SpanishMessages = map[string]string{
	"required":    "{field} es requerido",
	"not_empty":   "{field} no debe estar vacío",
	"email":       "{field} debe ser una dirección de correo válida",
	"url":         "{field} debe ser una URL válida",
	"uuid":        "{field} debe ser un UUID válido",
	"min":         "{field} debe ser al menos {min}",
	"max":         "{field} debe ser como máximo {max}",
	"min_length":  "{field} debe tener al menos {min} caracteres",
	"max_length":  "{field} debe tener como máximo {max} caracteres",
	"min_items":   "{field} debe tener al menos {min} elementos",
	"max_items":   "{field} debe tener como máximo {max} elementos",
	"pattern":     "{field} no coincide con el patrón requerido",
	"in":          "{field} debe ser uno de los valores permitidos",
	"not_in":      "{field} no debe ser uno de los valores prohibidos",
	"each":        "{field}[{index}]: {error}",
	"unique":      "{field} contiene valores duplicados",
	"unique_by":   "{field} contiene valores duplicados para la clave",
	"custom":      "{field} es inválido",
	"deferred":    "{field} validación fallida",
	"conversion":  "{field}: no se puede convertir el valor",
	"type":        "{field}: tipo esperado {expected}, recibido {actual}",
	"transform":   "{field}: error de transformación",
	"convert":     "{field}: error de conversión",
	"required_if": "{field} es requerido",
	"negative":    "{field} debe ser negativo",
	"positive":    "{field} debe ser positivo",
	"non_zero":    "{field} no debe ser cero",
}

// GermanMessages contains German translations.
var GermanMessages = map[string]string{
	"required":    "{field} ist erforderlich",
	"not_empty":   "{field} darf nicht leer sein",
	"email":       "{field} muss eine gültige E-Mail-Adresse sein",
	"url":         "{field} muss eine gültige URL sein",
	"uuid":        "{field} muss eine gültige UUID sein",
	"min":         "{field} muss mindestens {min} sein",
	"max":         "{field} darf höchstens {max} sein",
	"min_length":  "{field} muss mindestens {min} Zeichen haben",
	"max_length":  "{field} darf höchstens {max} Zeichen haben",
	"min_items":   "{field} muss mindestens {min} Elemente haben",
	"max_items":   "{field} darf höchstens {max} Elemente haben",
	"pattern":     "{field} entspricht nicht dem erforderlichen Muster",
	"in":          "{field} muss einer der erlaubten Werte sein",
	"not_in":      "{field} darf nicht einer der verbotenen Werte sein",
	"each":        "{field}[{index}]: {error}",
	"unique":      "{field} enthält doppelte Werte",
	"unique_by":   "{field} enthält doppelte Werte für den Schlüssel",
	"custom":      "{field} ist ungültig",
	"deferred":    "{field} Validierung fehlgeschlagen",
	"conversion":  "{field}: Wert kann nicht konvertiert werden",
	"type":        "{field}: erwarteter Typ {expected}, erhalten {actual}",
	"transform":   "{field}: Transformationsfehler",
	"convert":     "{field}: Konvertierungsfehler",
	"required_if": "{field} ist erforderlich",
	"negative":    "{field} muss negativ sein",
	"positive":    "{field} muss positiv sein",
	"non_zero":    "{field} darf nicht null sein",
}

// NewFrenchCatalog creates a message catalog with French translations.
func NewFrenchCatalog() *MessageCatalog {
	return &MessageCatalog{
		printer:  message.NewPrinter(language.French),
		messages: FrenchMessages,
	}
}

// NewSpanishCatalog creates a message catalog with Spanish translations.
func NewSpanishCatalog() *MessageCatalog {
	return &MessageCatalog{
		printer:  message.NewPrinter(language.Spanish),
		messages: SpanishMessages,
	}
}

// NewGermanCatalog creates a message catalog with German translations.
func NewGermanCatalog() *MessageCatalog {
	return &MessageCatalog{
		printer:  message.NewPrinter(language.German),
		messages: GermanMessages,
	}
}

// CatalogForLanguage returns a message catalog for the given language tag.
// Falls back to English if the language is not supported.
func CatalogForLanguage(tag language.Tag) *MessageCatalog {
	base, _ := tag.Base()
	switch base.String() {
	case "fr":
		return NewFrenchCatalog()
	case "es":
		return NewSpanishCatalog()
	case "de":
		return NewGermanCatalog()
	default:
		return defaultCatalog
	}
}
