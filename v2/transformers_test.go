package poxxy

import (
	"errors"
	"strings"
	"testing"
)

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no spaces", "hello", "hello"},
		{"leading spaces", "  hello", "hello"},
		{"trailing spaces", "hello  ", "hello"},
		{"both sides", "  hello  ", "hello"},
		{"tabs and newlines", "\t\nhello\t\n", "hello"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
	}

	tr := TrimSpace()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("TrimSpace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"all upper", "HELLO", "hello"},
		{"mixed case", "HeLLo WoRLD", "hello world"},
		{"already lower", "hello", "hello"},
		{"empty string", "", ""},
		{"with numbers", "Hello123", "hello123"},
	}

	tr := ToLower()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("ToLower(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"all lower", "hello", "HELLO"},
		{"mixed case", "HeLLo WoRLD", "HELLO WORLD"},
		{"already upper", "HELLO", "HELLO"},
		{"empty string", "", ""},
		{"with numbers", "Hello123", "HELLO123"},
	}

	tr := ToUpper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("ToUpper(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"all lower", "hello world", "Hello World"},
		{"all upper", "HELLO WORLD", "Hello World"},
		{"mixed case", "hELLO wORLD", "Hello World"},
		{"empty string", "", ""},
		{"single word", "hello", "Hello"},
	}

	tr := TitleCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("TitleCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"all lower", "hello", "Hello"},
		{"already capitalized", "Hello", "Hello"},
		{"all upper", "HELLO", "HELLO"},
		{"empty string", "", ""},
		{"single char", "a", "A"},
		{"sentence", "hello world", "Hello world"},
	}

	tr := Capitalize()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Capitalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal email", "test@example.com", "test@example.com"},
		{"uppercase", "TEST@EXAMPLE.COM", "test@example.com"},
		{"with spaces", "  test@example.com  ", "test@example.com"},
		{"mixed", "  TEST@Example.COM  ", "test@example.com"},
		{"empty string", "", ""},
	}

	tr := SanitizeEmail()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("SanitizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTrimPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		input    string
		expected string
	}{
		{"has prefix", "pre_", "pre_value", "value"},
		{"no prefix", "pre_", "value", "value"},
		{"empty input", "pre_", "", ""},
		{"prefix only", "pre_", "pre_", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := TrimPrefix(tt.prefix)
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("TrimPrefix(%q)(%q) = %q, want %q", tt.prefix, tt.input, result, tt.expected)
			}
		})
	}
}

func TestTrimSuffix(t *testing.T) {
	tests := []struct {
		name     string
		suffix   string
		input    string
		expected string
	}{
		{"has suffix", "_suf", "value_suf", "value"},
		{"no suffix", "_suf", "value", "value"},
		{"empty input", "_suf", "", ""},
		{"suffix only", "_suf", "_suf", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := TrimSuffix(tt.suffix)
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("TrimSuffix(%q)(%q) = %q, want %q", tt.suffix, tt.input, result, tt.expected)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		input    string
		expected string
	}{
		{"single replace", " ", "-", "hello world", "hello-world"},
		{"multiple replace", "a", "x", "banana", "bxnxnx"},
		{"no match", "x", "y", "hello", "hello"},
		{"empty input", " ", "-", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Replace(tt.old, tt.new)
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Replace(%q, %q)(%q) = %q, want %q", tt.old, tt.new, tt.input, result, tt.expected)
			}
		})
	}
}

func TestReplaceN(t *testing.T) {
	tr := ReplaceN("a", "x", 2)
	result, err := tr.Transform("banana")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "bxnxna"
	if result != expected {
		t.Errorf("ReplaceN('a', 'x', 2)('banana') = %q, want %q", result, expected)
	}
}

func TestAbs(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tr := Abs[int]()
		tests := []struct {
			input    int
			expected int
		}{
			{5, 5},
			{-5, 5},
			{0, 0},
		}
		for _, tt := range tests {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Abs(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("float64", func(t *testing.T) {
		tr := Abs[float64]()
		tests := []struct {
			input    float64
			expected float64
		}{
			{5.5, 5.5},
			{-5.5, 5.5},
			{0.0, 0.0},
		}
		for _, tt := range tests {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Abs(%f) = %f, want %f", tt.input, result, tt.expected)
			}
		}
	})
}

func TestClamp(t *testing.T) {
	tr := Clamp(0, 100)

	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"within range", 50, 50},
		{"at min", 0, 0},
		{"at max", 100, 100},
		{"below min", -10, 0},
		{"above max", 150, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Clamp(0, 100)(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultIfZero(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		tr := DefaultIfZero("default")
		tests := []struct {
			input    string
			expected string
		}{
			{"", "default"},
			{"value", "value"},
		}
		for _, tt := range tests {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("DefaultIfZero(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("int", func(t *testing.T) {
		tr := DefaultIfZero(42)
		tests := []struct {
			input    int
			expected int
		}{
			{0, 42},
			{10, 10},
		}
		for _, tt := range tests {
			result, err := tr.Transform(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("DefaultIfZero(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	})
}

func TestMapSlice(t *testing.T) {
	tr := MapSlice(func(s string) string {
		return strings.ToUpper(s)
	})

	input := []string{"hello", "world"}
	result, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"HELLO", "WORLD"}
	if len(result) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(result), len(expected))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, result[i], expected[i])
		}
	}
}

func TestFilterSlice(t *testing.T) {
	tr := FilterSlice(func(n int) bool {
		return n > 0
	})

	input := []int{-2, -1, 0, 1, 2, 3}
	result, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []int{1, 2, 3}
	if len(result) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(result), len(expected))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("result[%d] = %d, want %d", i, result[i], expected[i])
		}
	}
}

func TestCompact(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		tr := Compact[string]()
		input := []string{"hello", "", "world", "", "!"}
		result, err := tr.Transform(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"hello", "world", "!"}
		if len(result) != len(expected) {
			t.Fatalf("length mismatch: got %d, want %d", len(result), len(expected))
		}
		for i := range result {
			if result[i] != expected[i] {
				t.Errorf("result[%d] = %q, want %q", i, result[i], expected[i])
			}
		}
	})

	t.Run("ints", func(t *testing.T) {
		tr := Compact[int]()
		input := []int{1, 0, 2, 0, 3}
		result, err := tr.Transform(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []int{1, 2, 3}
		if len(result) != len(expected) {
			t.Fatalf("length mismatch: got %d, want %d", len(result), len(expected))
		}
	})
}

func TestTransformerFunc(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tr := TransformerFunc(func(s string) (string, error) {
			return strings.ToUpper(s), nil
		})

		result, err := tr.Transform("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "HELLO" {
			t.Errorf("got %q, want %q", result, "HELLO")
		}
	})

	t.Run("error", func(t *testing.T) {
		tr := TransformerFunc(func(s string) (string, error) {
			if s == "" {
				return "", errors.New("empty string not allowed")
			}
			return s, nil
		})

		_, err := tr.Transform("")
		if err == nil {
			t.Error("expected error for empty string")
		}
	})

	t.Run("name", func(t *testing.T) {
		tr := TransformerFunc(func(s string) (string, error) {
			return s, nil
		})
		if tr.Name() != "Custom" {
			t.Errorf("expected name 'Custom', got %q", tr.Name())
		}
	})
}

func TestTransformerFuncNamed(t *testing.T) {
	tr := TransformerFuncNamed("Slugify", func(s string) (string, error) {
		return strings.ReplaceAll(strings.ToLower(s), " ", "-"), nil
	})

	if tr.Name() != "Slugify" {
		t.Errorf("expected name 'Slugify', got %q", tr.Name())
	}

	result, err := tr.Transform("Hello World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello-world" {
		t.Errorf("got %q, want %q", result, "hello-world")
	}
}

func TestChain(t *testing.T) {
	tr := Chain(
		TrimSpace(),
		ToLower(),
		Replace(" ", "-"),
	)

	result, err := tr.Transform("  Hello World  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "hello-world"
	if result != expected {
		t.Errorf("Chain() = %q, want %q", result, expected)
	}
}

func TestTransformerName(t *testing.T) {
	tests := []struct {
		name     string
		tr       Transformer[string]
		expected string
	}{
		{"TrimSpace", TrimSpace(), "TrimSpace"},
		{"ToLower", ToLower(), "ToLower"},
		{"ToUpper", ToUpper(), "ToUpper"},
		{"TitleCase", TitleCase(), "TitleCase"},
		{"Capitalize", Capitalize(), "Capitalize"},
		{"SanitizeEmail", SanitizeEmail(), "SanitizeEmail"},
		{"TrimPrefix", TrimPrefix("x"), "TrimPrefix"},
		{"TrimSuffix", TrimSuffix("x"), "TrimSuffix"},
		{"Replace", Replace("a", "b"), "Replace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tr.Name() != tt.expected {
				t.Errorf("expected name %q, got %q", tt.expected, tt.tr.Name())
			}
		})
	}
}

func TestIntegrationTransformersWithSchema(t *testing.T) {
	var email string
	var name string
	var tags []string

	schema := NewSchema(
		Field("email", &email).
			Transform(TrimSpace(), SanitizeEmail()).
			Validate(Required[string](), Email()),
		Field("name", &name).
			Transform(TrimSpace(), TitleCase()),
		Slice("tags", &tags).
			Validate(MinItems[string](1)),
	)

	input := `{"email": "  TEST@EXAMPLE.COM  ", "name": "  john doe  ", "tags": ["go"]}`
	err := schema.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email != "test@example.com" {
		t.Errorf("expected email='test@example.com', got %q", email)
	}
	if name != "John Doe" {
		t.Errorf("expected name='John Doe', got %q", name)
	}
}
