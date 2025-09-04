package poxxy

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestTransformers_EdgeCases(t *testing.T) {
	t.Run("ToUpper with unicode", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(ToUpper())),
		)

		unicodeCases := []struct {
			input    string
			expected string
		}{
			{"café", "CAFÉ"},
			{"naïve", "NAÏVE"},
			{"über", "ÜBER"},
			{"ß", "ß"},   // German sharp S - stays the same in uppercase
			{"你好", "你好"}, // Chinese characters
		}

		for _, tc := range unicodeCases {
			err := schema.Apply(map[string]interface{}{"test": tc.input})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		}
	})

	t.Run("ToLower with special characters", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(ToLower())),
		)

		specialCases := []struct {
			input    string
			expected string
		}{
			{"CAFÉ", "café"},
			{"NAÏVE", "naïve"},
			{"ÜBER", "über"},
			{"SS", "ss"},
			{"HELLO WORLD", "hello world"},
		}

		for _, tc := range specialCases {
			err := schema.Apply(map[string]interface{}{"test": tc.input})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		}
	})

	t.Run("TrimSpace with various whitespace", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(TrimSpace())),
		)

		whitespaceCases := []struct {
			input    string
			expected string
		}{
			{"  hello  ", "hello"},
			{"\t\n\r hello \t\n\r", "hello"},
			{"\u00A0hello\u00A0", "hello"}, // Non-breaking space
			{"\u2000hello\u2000", "hello"}, // En quad
			{"\u2001hello\u2001", "hello"}, // Em quad
		}

		for _, tc := range whitespaceCases {
			err := schema.Apply(map[string]interface{}{"test": tc.input})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		}
	})

	t.Run("TitleCase with edge cases", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(TitleCase())),
		)

		titleCases := []struct {
			input    string
			expected string
		}{
			{"hello world", "Hello World"},
			{"", ""},
			{"a", "A"},
			{"hello-world", "Hello-World"},
			{"hello_world", "Hello_world"},
			{"hello.world", "Hello.world"},
		}

		for _, tc := range titleCases {
			value = "" // reset value to empty string
			err := schema.Apply(map[string]interface{}{"test": tc.input})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		}
	})

	t.Run("Capitalize with edge cases", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(Capitalize())),
		)

		capitalizeCases := []struct {
			input    string
			expected string
		}{
			{"hello", "Hello"},
			{"", ""},
			{"a", "A"},
			{"HELLO", "Hello"},
			{"hELLO", "Hello"},
			{"123hello", "123hello"}, // Numbers at start
			{"!hello", "!hello"},     // Special chars at start
		}

		for _, tc := range capitalizeCases {
			value = "" // reset value to empty string
			err := schema.Apply(map[string]interface{}{"test": tc.input})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		}
	})

	t.Run("SanitizeEmail with edge cases", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(SanitizeEmail())),
		)

		emailCases := []struct {
			input    string
			expected string
		}{
			{"  TEST@EXAMPLE.COM  ", "test@example.com"},
			{"Test@Example.Com", "test@example.com"},
			{"test@example.com", "test@example.com"},
			{"", ""},
			{"   ", ""},
		}

		for _, tc := range emailCases {
			value = "" // reset value to empty string
			err := schema.Apply(map[string]interface{}{"test": tc.input})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		}
	})

	t.Run("multiple transformers order", func(t *testing.T) {
		var value string
		schema := NewSchema(
			Value("test", &value, WithTransformers(
				TrimSpace(),
				ToLower(),
				Capitalize(),
			)),
		)

		err := schema.Apply(map[string]interface{}{"test": "  HELLO WORLD  "})
		assert.NoError(t, err)
		assert.Equal(t, "Hello world", value)
	})

	t.Run("transformer with error", func(t *testing.T) {
		var value string
		errorTransformer := CustomTransformer[string](func(s string) (string, error) {
			if strings.Contains(s, "error") {
				return "", fmt.Errorf("transformation failed")
			}
			return strings.ToUpper(s), nil
		})

		schema := NewSchema(
			Value("test", &value, WithTransformers(errorTransformer)),
		)

		// Test with error
		err := schema.Apply(map[string]interface{}{"test": "this has error"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transformation failed")

		// Test without error
		err = schema.Apply(map[string]interface{}{"test": "hello"})
		assert.NoError(t, err)
		assert.Equal(t, "HELLO", value)
	})

	t.Run("transformer with panic", func(t *testing.T) {
		var value string
		panicTransformer := CustomTransformer[string](func(s string) (string, error) {
			if s == "panic" {
				panic("transformer panic")
			}
			return s, nil
		})

		schema := NewSchema(
			Value("test", &value, WithTransformers(panicTransformer)),
		)

		// The panic should propagate
		assert.Panics(t, func() {
			schema.Apply(map[string]interface{}{"test": "panic"})
		})
	})

	t.Run("transformer with pointer field", func(t *testing.T) {
		var value *string
		upperTransformer := CustomTransformer[string](func(s string) (string, error) {
			return strings.ToUpper(s), nil
		})

		schema := NewSchema(
			Pointer("test", &value, WithTransformers(upperTransformer)),
		)

		// Test with valid value - transformer should be called
		err := schema.Apply(map[string]interface{}{"test": "hello"})
		assert.NoError(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, "HELLO", *value)
	})

	t.Run("transformer with empty string", func(t *testing.T) {
		var value string
		emptyTransformer := CustomTransformer[string](func(s string) (string, error) {
			if s == "" {
				return "default", nil
			}
			return s, nil
		})

		schema := NewSchema(
			Value("test", &value, WithTransformers(emptyTransformer)),
		)

		// Empty strings don't go through transformers in ValueField
		err := schema.Apply(map[string]interface{}{"test": ""})
		assert.NoError(t, err)
		assert.Equal(t, "", value) // Empty string remains empty

		// Test with non-empty string
		err = schema.Apply(map[string]interface{}{"test": "hello"})
		assert.NoError(t, err)
		assert.Equal(t, "hello", value)
	})

	t.Run("transformer with very long string", func(t *testing.T) {
		var value string
		longTransformer := CustomTransformer[string](func(s string) (string, error) {
			if len(s) > 1000000 {
				return "", fmt.Errorf("string too long")
			}
			return strings.ToUpper(s), nil
		})

		schema := NewSchema(
			Value("test", &value, WithTransformers(longTransformer)),
		)

		// Test with very long string
		longString := strings.Repeat("a", 1000001)
		err := schema.Apply(map[string]interface{}{"test": longString})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "string too long")

		// Test with normal string
		err = schema.Apply(map[string]interface{}{"test": "hello"})
		assert.NoError(t, err)
		assert.Equal(t, "HELLO", value)
	})

	t.Run("transformer with special unicode", func(t *testing.T) {
		var value string
		unicodeTransformer := CustomTransformer[string](func(s string) (string, error) {
			// Transformer that handles special characters
			var result strings.Builder
			for _, r := range s {
				if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
					result.WriteRune(r)
				} else {
					result.WriteRune('_')
				}
			}
			return result.String(), nil
		})

		schema := NewSchema(
			Value("test", &value, WithTransformers(unicodeTransformer)),
		)

		specialCases := []struct {
			input    string
			expected string
		}{
			{"hello@world.com", "hello_world_com"},
			{"file-name.txt", "file_name_txt"},
			{"user (admin)", "user _admin_"},
			{"price: $100", "price_ _100"},
		}

		for _, tc := range specialCases {
			err := schema.Apply(map[string]interface{}{"test": tc.input})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		}
	})
}
