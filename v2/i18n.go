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
	"email":       "{field} must be a valid email address",
	"url":         "{field} must be a valid URL",
	"uuid":        "{field} must be a valid UUID",
	"min":         "{field} must be at least {min}",
	"max":         "{field} must be at most {max}",
	"minlength":   "{field} must be at least {min} characters",
	"maxlength":   "{field} must be at most {max} characters",
	"minitems":    "{field} must have at least {min} items",
	"maxitems":    "{field} must have at most {max} items",
	"pattern":     "{field} does not match the required pattern",
	"in":          "{field} must be one of the allowed values",
	"each":        "{field}[{index}]: {error}",
	"unique":      "{field} contains duplicate values",
	"uniqueby":    "{field} contains duplicate values for the key",
	"custom":      "{field} is invalid",
	"conversion":  "{field}: cannot convert value",
	"type":        "{field}: expected type {expected}, got {actual}",
	"negative":    "{field} must be negative",
	"positive":    "{field} must be positive",
	"nonzero":     "{field} must not be zero",
	"notempty":    "{field} must not be empty",
}

// =============================================================================
// Message Key Type
// =============================================================================

// MessageKey is a key for looking up translated messages.
type MessageKey string

// Common message keys for built-in validators.
const (
	MsgRequired   MessageKey = "required"
	MsgEmail      MessageKey = "email"
	MsgURL        MessageKey = "url"
	MsgUUID       MessageKey = "uuid"
	MsgMin        MessageKey = "min"
	MsgMax        MessageKey = "max"
	MsgMinLength  MessageKey = "minlength"
	MsgMaxLength  MessageKey = "maxlength"
	MsgMinItems   MessageKey = "minitems"
	MsgMaxItems   MessageKey = "maxitems"
	MsgPattern    MessageKey = "pattern"
	MsgIn         MessageKey = "in"
	MsgEach       MessageKey = "each"
	MsgUnique     MessageKey = "unique"
	MsgUniqueBy   MessageKey = "uniqueby"
	MsgCustom     MessageKey = "custom"
	MsgConversion MessageKey = "conversion"
	MsgType       MessageKey = "type"
	MsgNegative   MessageKey = "negative"
	MsgPositive   MessageKey = "positive"
	MsgNonZero    MessageKey = "nonzero"
	MsgNotEmpty   MessageKey = "notempty"
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
	"required":   "{field} est requis",
	"email":      "{field} doit être une adresse email valide",
	"url":        "{field} doit être une URL valide",
	"uuid":       "{field} doit être un UUID valide",
	"min":        "{field} doit être au moins {min}",
	"max":        "{field} doit être au plus {max}",
	"minlength":  "{field} doit contenir au moins {min} caractères",
	"maxlength":  "{field} doit contenir au plus {max} caractères",
	"minitems":   "{field} doit contenir au moins {min} éléments",
	"maxitems":   "{field} doit contenir au plus {max} éléments",
	"pattern":    "{field} ne correspond pas au format requis",
	"in":         "{field} doit être l'une des valeurs autorisées",
	"each":       "{field}[{index}]: {error}",
	"unique":     "{field} contient des doublons",
	"uniqueby":   "{field} contient des doublons pour la clé",
	"custom":     "{field} est invalide",
	"conversion": "{field}: impossible de convertir la valeur",
	"type":       "{field}: type attendu {expected}, reçu {actual}",
	"negative":   "{field} doit être négatif",
	"positive":   "{field} doit être positif",
	"nonzero":    "{field} ne doit pas être zéro",
	"notempty":   "{field} ne doit pas être vide",
}

// SpanishMessages contains Spanish translations.
var SpanishMessages = map[string]string{
	"required":   "{field} es requerido",
	"email":      "{field} debe ser una dirección de correo válida",
	"url":        "{field} debe ser una URL válida",
	"uuid":       "{field} debe ser un UUID válido",
	"min":        "{field} debe ser al menos {min}",
	"max":        "{field} debe ser como máximo {max}",
	"minlength":  "{field} debe tener al menos {min} caracteres",
	"maxlength":  "{field} debe tener como máximo {max} caracteres",
	"minitems":   "{field} debe tener al menos {min} elementos",
	"maxitems":   "{field} debe tener como máximo {max} elementos",
	"pattern":    "{field} no coincide con el patrón requerido",
	"in":         "{field} debe ser uno de los valores permitidos",
	"each":       "{field}[{index}]: {error}",
	"unique":     "{field} contiene valores duplicados",
	"uniqueby":   "{field} contiene valores duplicados para la clave",
	"custom":     "{field} es inválido",
	"conversion": "{field}: no se puede convertir el valor",
	"type":       "{field}: tipo esperado {expected}, recibido {actual}",
	"negative":   "{field} debe ser negativo",
	"positive":   "{field} debe ser positivo",
	"nonzero":    "{field} no debe ser cero",
	"notempty":   "{field} no debe estar vacío",
}

// GermanMessages contains German translations.
var GermanMessages = map[string]string{
	"required":   "{field} ist erforderlich",
	"email":      "{field} muss eine gültige E-Mail-Adresse sein",
	"url":        "{field} muss eine gültige URL sein",
	"uuid":       "{field} muss eine gültige UUID sein",
	"min":        "{field} muss mindestens {min} sein",
	"max":        "{field} darf höchstens {max} sein",
	"minlength":  "{field} muss mindestens {min} Zeichen haben",
	"maxlength":  "{field} darf höchstens {max} Zeichen haben",
	"minitems":   "{field} muss mindestens {min} Elemente haben",
	"maxitems":   "{field} darf höchstens {max} Elemente haben",
	"pattern":    "{field} entspricht nicht dem erforderlichen Muster",
	"in":         "{field} muss einer der erlaubten Werte sein",
	"each":       "{field}[{index}]: {error}",
	"unique":     "{field} enthält doppelte Werte",
	"uniqueby":   "{field} enthält doppelte Werte für den Schlüssel",
	"custom":     "{field} ist ungültig",
	"conversion": "{field}: Wert kann nicht konvertiert werden",
	"type":       "{field}: erwarteter Typ {expected}, erhalten {actual}",
	"negative":   "{field} muss negativ sein",
	"positive":   "{field} muss positiv sein",
	"nonzero":    "{field} darf nicht null sein",
	"notempty":   "{field} darf nicht leer sein",
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
