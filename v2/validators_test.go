package poxxy

import (
	"errors"
	"strings"
	"testing"
)

func TestRequired(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"whitespace only", "   ", false}, // Not empty, use NotEmpty for this
	}

	v := Required[string]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("Required().Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequiredInt(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"zero", 0, true},
		{"positive", 42, false},
		{"negative", -1, false},
	}

	v := Required[int]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("Required[int]().Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotEmpty(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"whitespace only", "   ", false}, // Has length > 0
	}

	v := NotEmpty[string]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("NotEmpty().Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty", "", false}, // Empty is valid, use Required() for non-empty
		{"valid email", "test@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"missing @", "testexample.com", true},
		{"missing domain", "test@", true},
		{"missing local part", "@example.com", true},
		{"spaces", "test @example.com", true},
		{"double @", "test@@example.com", true},
	}

	v := Email()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.value, "email")
			if (err != nil) != tt.wantErr {
				t.Errorf("Email().Validate(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty", "", false},
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"with path", "https://example.com/path/to/page", false},
		{"with query", "https://example.com?query=value", false},
		{"missing protocol", "example.com", true},
		{"ftp protocol", "ftp://example.com", true},
		{"just protocol", "http://", true},
	}

	v := URL()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.value, "url")
			if (err != nil) != tt.wantErr {
				t.Errorf("URL().Validate(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		value   string
		wantErr bool
	}{
		{"empty value", `^\d+$`, "", false},
		{"digits only valid", `^\d+$`, "12345", false},
		{"digits only invalid", `^\d+$`, "123abc", true},
		{"zip code valid", `^\d{5}$`, "12345", false},
		{"zip code invalid", `^\d{5}$`, "1234", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Pattern(tt.pattern)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("Pattern(%q).Validate(%q) error = %v, wantErr %v", tt.pattern, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		value   int
		wantErr bool
	}{
		{"equal to min", 5, 5, false},
		{"greater than min", 5, 10, false},
		{"less than min", 5, 3, true},
		{"negative values", -10, -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Min(tt.min)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("Min(%d).Validate(%d) error = %v, wantErr %v", tt.min, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name    string
		max     int
		value   int
		wantErr bool
	}{
		{"equal to max", 10, 10, false},
		{"less than max", 10, 5, false},
		{"greater than max", 10, 15, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Max(tt.max)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("Max(%d).Validate(%d) error = %v, wantErr %v", tt.max, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestBetween(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		max     int
		value   int
		wantErr bool
	}{
		{"equal to min", 5, 10, 5, false},
		{"equal to max", 5, 10, 10, false},
		{"in range", 5, 10, 7, false},
		{"below min", 5, 10, 3, true},
		{"above max", 5, 10, 15, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Between(tt.min, tt.max)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("Between(%d, %d).Validate(%d) error = %v, wantErr %v", tt.min, tt.max, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		value   string
		wantErr bool
	}{
		{"exact length", 3, "abc", false},
		{"longer", 3, "abcdef", false},
		{"shorter", 3, "ab", true},
		{"empty", 1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := MinLength[string](tt.min)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("MinLength(%d).Validate(%q) error = %v, wantErr %v", tt.min, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name    string
		max     int
		value   string
		wantErr bool
	}{
		{"exact length", 5, "hello", false},
		{"shorter", 5, "hi", false},
		{"longer", 5, "hello world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := MaxLength[string](tt.max)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("MaxLength(%d).Validate(%q) error = %v, wantErr %v", tt.max, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		value   string
		wantErr bool
	}{
		{"exact length", 5, "hello", false},
		{"shorter", 5, "hi", true},
		{"longer", 5, "hello world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Length[string](tt.length)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("Length(%d).Validate(%q) error = %v, wantErr %v", tt.length, tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestMinItems(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		value   []int
		wantErr bool
	}{
		{"exact count", 2, []int{1, 2}, false},
		{"more items", 2, []int{1, 2, 3}, false},
		{"fewer items", 2, []int{1}, true},
		{"empty", 1, []int{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := MinItems[int](tt.min)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("MinItems(%d).Validate() error = %v, wantErr %v", tt.min, err, tt.wantErr)
			}
		})
	}
}

func TestMaxItems(t *testing.T) {
	tests := []struct {
		name    string
		max     int
		value   []int
		wantErr bool
	}{
		{"exact count", 3, []int{1, 2, 3}, false},
		{"fewer items", 3, []int{1, 2}, false},
		{"more items", 3, []int{1, 2, 3, 4}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := MaxItems[int](tt.max)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("MaxItems(%d).Validate() error = %v, wantErr %v", tt.max, err, tt.wantErr)
			}
		})
	}
}

func TestEach(t *testing.T) {
	t.Run("all valid", func(t *testing.T) {
		v := Each(Min(0), Max(100))
		err := v.Validate([]int{10, 20, 30}, "scores")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("one invalid", func(t *testing.T) {
		v := Each(Min(0), Max(100))
		err := v.Validate([]int{10, 150, 30}, "scores")
		if err == nil {
			t.Error("expected error for value > 100")
		}
	})

	t.Run("string validation", func(t *testing.T) {
		v := Each(MinLength[string](2))
		err := v.Validate([]string{"ab", "abc", "a"}, "names")
		if err == nil {
			t.Error("expected error for single character")
		}
	})
}

func TestUnique(t *testing.T) {
	t.Run("all unique", func(t *testing.T) {
		v := Unique[int]()
		err := v.Validate([]int{1, 2, 3, 4, 5}, "ids")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("with duplicates", func(t *testing.T) {
		v := Unique[int]()
		err := v.Validate([]int{1, 2, 3, 2, 5}, "ids")
		if err == nil {
			t.Error("expected error for duplicate")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		v := Unique[string]()
		err := v.Validate([]string{}, "tags")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestUniqueBy(t *testing.T) {
	type User struct {
		ID    int
		Email string
	}

	t.Run("unique by key", func(t *testing.T) {
		v := UniqueBy(func(u User) string { return u.Email })
		users := []User{
			{ID: 1, Email: "a@example.com"},
			{ID: 2, Email: "b@example.com"},
		}
		err := v.Validate(users, "users")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("duplicate key", func(t *testing.T) {
		v := UniqueBy(func(u User) string { return u.Email })
		users := []User{
			{ID: 1, Email: "a@example.com"},
			{ID: 2, Email: "a@example.com"}, // duplicate email
		}
		err := v.Validate(users, "users")
		if err == nil {
			t.Error("expected error for duplicate email")
		}
	})
}

func TestIn(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		value   string
		wantErr bool
	}{
		{"value in list", []string{"a", "b", "c"}, "b", false},
		{"value not in list", []string{"a", "b", "c"}, "d", true},
		{"empty value in list", []string{"", "a", "b"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := In(tt.allowed...)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("In().Validate(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestNotIn(t *testing.T) {
	tests := []struct {
		name       string
		disallowed []string
		value      string
		wantErr    bool
	}{
		{"value in blacklist", []string{"admin", "root"}, "admin", true},
		{"value not in blacklist", []string{"admin", "root"}, "user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NotIn(tt.disallowed...)
			err := v.Validate(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Errorf("NotIn().Validate(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidatorFunc(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		v := ValidatorFunc(func(s string) error {
			if !strings.HasPrefix(s, "prefix_") {
				return errors.New("must start with prefix_")
			}
			return nil
		})

		err := v.Validate("prefix_test", "field")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		v := ValidatorFunc(func(s string) error {
			if !strings.HasPrefix(s, "prefix_") {
				return errors.New("must start with prefix_")
			}
			return nil
		})

		err := v.Validate("no_prefix", "field")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("with custom message", func(t *testing.T) {
		v := ValidatorFunc(func(s string) error {
			return errors.New("original error")
		}).WithMessage("Custom error message")

		err := v.Validate("value", "field")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "Custom error message") {
			t.Errorf("expected custom message, got: %v", err)
		}
	})
}

func TestWithMessage(t *testing.T) {
	v := Required[string]().WithMessage("{field} cannot be blank")
	err := v.Validate("", "username")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "username cannot be blank") {
		t.Errorf("expected interpolated message, got: %v", err)
	}
}

func TestValidatorRule(t *testing.T) {
	tests := []struct {
		name     string
		v        Validator[string]
		expected string
	}{
		{"required", Required[string](), "required"},
		{"email", Email(), "email"},
		{"url", URL(), "url"},
		{"pattern", Pattern(`\d+`), "pattern"},
		{"min_length", MinLength[string](5), "min_length"},
		{"max_length", MaxLength[string](10), "max_length"},
		{"in", In("a", "b"), "in"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.v.Rule() != tt.expected {
				t.Errorf("expected rule %q, got %q", tt.expected, tt.v.Rule())
			}
		})
	}
}

func TestIntegrationWithSchema(t *testing.T) {
	var email string
	var age int
	var tags []string

	schema := NewSchema(
		Field("email", &email).
			Validate(Required[string](), Email()),
		Field("age", &age).
			Validate(Required[int](), Min(0), Max(150)),
		Slice("tags", &tags).
			Validate(MinItems[string](1), MaxItems[string](5)),
	)

	t.Run("valid data", func(t *testing.T) {
		input := `{"email": "test@example.com", "age": 25, "tags": ["go", "validation"]}`
		err := schema.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if email != "test@example.com" {
			t.Errorf("expected email='test@example.com', got %q", email)
		}
		if age != 25 {
			t.Errorf("expected age=25, got %d", age)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		email = ""
		input := `{"email": "invalid-email", "age": 25, "tags": ["go"]}`
		err := schema.Parse(strings.NewReader(input))
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("age out of range", func(t *testing.T) {
		email = ""
		age = 0
		input := `{"email": "test@example.com", "age": 200, "tags": ["go"]}`
		err := schema.Parse(strings.NewReader(input))
		if err == nil {
			t.Fatal("expected validation error for age > 150")
		}
	})

	t.Run("empty tags", func(t *testing.T) {
		email = ""
		age = 0
		tags = nil
		input := `{"email": "test@example.com", "age": 25, "tags": []}`
		err := schema.Parse(strings.NewReader(input))
		if err == nil {
			t.Fatal("expected validation error for empty tags")
		}
	})
}

func TestUUID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty string", "", false}, // Empty is allowed, use Required for non-empty
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uuid uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"invalid - no dashes", "550e8400e29b41d4a716446655440000", true},
		{"invalid - wrong format", "550e8400-e29b-41d4-a716", true},
		{"invalid - not hex", "gggg8400-e29b-41d4-a716-446655440000", true},
		{"random string", "not-a-uuid", true},
	}

	v := UUID()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.value, "id")
			if (err != nil) != tt.wantErr {
				t.Errorf("UUID().Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeferredValidator(t *testing.T) {
	// Simulate a database check
	existingEmails := map[string]bool{
		"taken@example.com": true,
	}

	emailChecker := Deferred[string](func(email string, field string) DeferredCheck {
		return func() error {
			if existingEmails[email] {
				return errors.New("email already registered")
			}
			return nil
		}
	})

	t.Run("valid deferred check", func(t *testing.T) {
		check, err := emailChecker.Validate("new@example.com", "email")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if check == nil {
			t.Fatal("expected check function")
		}
		// Run the deferred check
		if err := check(); err != nil {
			t.Fatalf("unexpected check error: %v", err)
		}
	})

	t.Run("failing deferred check", func(t *testing.T) {
		check, err := emailChecker.Validate("taken@example.com", "email")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Run the deferred check
		if err := check(); err == nil {
			t.Fatal("expected check to fail")
		}
	})
}

func TestDeferredValidatorNamed(t *testing.T) {
	checker := DeferredNamed[string]("email_unique", func(email string, field string) DeferredCheck {
		return func() error {
			return nil
		}
	})

	if checker.Rule() != "email_unique" {
		t.Errorf("expected rule 'email_unique', got %q", checker.Rule())
	}
}

func TestDeferredChecks_Run(t *testing.T) {
	checks := &DeferredChecks{}

	checks.Add("email", func() error {
		return nil
	})
	checks.Add("username", func() error {
		return errors.New("username taken")
	})
	checks.Add("phone", func() error {
		return nil
	})

	errs := checks.Run()
	if errs == nil {
		t.Fatal("expected errors")
	}
	if len(errs.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs.Errors))
	}
	if errs.Errors[0].Field != "username" {
		t.Errorf("expected field 'username', got %q", errs.Errors[0].Field)
	}
}

func TestDeferredChecks_RunParallel(t *testing.T) {
	checks := &DeferredChecks{}

	// Add multiple checks that will run in parallel
	for i := 0; i < 5; i++ {
		idx := i
		checks.Add("field"+string(rune('0'+idx)), func() error {
			if idx == 2 || idx == 4 {
				return errors.New("failed")
			}
			return nil
		})
	}

	errs := checks.RunParallel()
	if errs == nil {
		t.Fatal("expected errors")
	}
	if len(errs.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs.Errors))
	}
}

func TestDeferredChecks_NoChecks(t *testing.T) {
	checks := &DeferredChecks{}

	if checks.HasChecks() {
		t.Error("expected no checks")
	}

	// Run should return nil with no checks
	if errs := checks.Run(); errs != nil {
		t.Errorf("expected nil, got %v", errs)
	}

	// RunParallel should also return nil with no checks
	if errs := checks.RunParallel(); errs != nil {
		t.Errorf("expected nil, got %v", errs)
	}
}

func TestDeferredChecks_AddNil(t *testing.T) {
	checks := &DeferredChecks{}
	checks.Add("field", nil) // Should not panic or add

	if checks.HasChecks() {
		t.Error("nil checks should not be added")
	}
}
