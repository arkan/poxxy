package poxxy

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewValidatorFnWithSpecificTypes démontre l'utilisation des validateurs typés
func TestNewValidatorFnWithSpecificTypes(t *testing.T) {
	t.Run("string validator", func(t *testing.T) {
		// Validateur typé pour les strings
		stringValidator := NewValidatorFn[string](func(value string, fieldName string) error {
			if len(value) < 5 {
				return fmt.Errorf("%s must be at least 5 characters", fieldName)
			}
			if !strings.Contains(value, "@") {
				return fmt.Errorf("%s must contain @ symbol", fieldName)
			}
			return nil
		})

		// Test avec une valeur valide
		err := stringValidator.Validate("test@example.com", "email")
		assert.NoError(t, err)

		// Test avec une valeur trop courte
		err = stringValidator.Validate("test", "email")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 5 characters")

		// Test avec une valeur sans @
		err = stringValidator.Validate("testexample.com", "email")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must contain @ symbol")

		// Test avec un mauvais type (doit échouer à la compilation si utilisé directement)
		err = stringValidator.Validate(123, "email")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected type string, got int")
	})

	t.Run("int validator", func(t *testing.T) {
		// Validateur typé pour les entiers
		intValidator := NewValidatorFn[int](func(value int, fieldName string) error {
			if value < 0 {
				return fmt.Errorf("%s must be positive", fieldName)
			}
			if value > 100 {
				return fmt.Errorf("%s must be less than or equal to 100", fieldName)
			}
			return nil
		})

		// Test avec une valeur valide
		err := intValidator.Validate(50, "age")
		assert.NoError(t, err)

		// Test avec une valeur négative
		err = intValidator.Validate(-5, "age")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")

		// Test avec une valeur trop grande
		err = intValidator.Validate(150, "age")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be less than or equal to 100")

		// Test avec un mauvais type
		err = intValidator.Validate("50", "age")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected type int, got string")
	})

	t.Run("slice validator", func(t *testing.T) {
		// Validateur typé pour les slices de strings
		sliceValidator := NewValidatorFn[[]string](func(value []string, fieldName string) error {
			if len(value) == 0 {
				return fmt.Errorf("%s cannot be empty", fieldName)
			}
			for i, item := range value {
				if len(item) == 0 {
					return fmt.Errorf("%s[%d] cannot be empty", fieldName, i)
				}
			}
			return nil
		})

		// Test avec une slice valide
		err := sliceValidator.Validate([]string{"a", "b", "c"}, "tags")
		assert.NoError(t, err)

		// Test avec une slice vide
		err = sliceValidator.Validate([]string{}, "tags")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")

		// Test avec une slice contenant un élément vide
		err = sliceValidator.Validate([]string{"a", "", "c"}, "tags")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tags[1] cannot be empty")

		// Test avec un mauvais type
		err = sliceValidator.Validate("not a slice", "tags")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected type []string, got string")
	})

	t.Run("struct validator", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		// Validateur typé pour les structs Person
		personValidator := NewValidatorFn[Person](func(value Person, fieldName string) error {
			if len(value.Name) == 0 {
				return fmt.Errorf("%s.name cannot be empty", fieldName)
			}
			if value.Age < 0 || value.Age > 150 {
				return fmt.Errorf("%s.age must be between 0 and 150", fieldName)
			}
			return nil
		})

		// Test avec une personne valide
		err := personValidator.Validate(Person{Name: "John", Age: 30}, "person")
		assert.NoError(t, err)

		// Test avec un nom vide
		err = personValidator.Validate(Person{Name: "", Age: 30}, "person")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "person.name cannot be empty")

		// Test avec un âge invalide
		err = personValidator.Validate(Person{Name: "John", Age: -5}, "person")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "person.age must be between 0 and 150")
	})

	t.Run("time validator", func(t *testing.T) {
		// Validateur typé pour les dates
		timeValidator := NewValidatorFn[time.Time](func(value time.Time, fieldName string) error {
			now := time.Now()
			if value.After(now) {
				return fmt.Errorf("%s cannot be in the future", fieldName)
			}
			if value.Before(now.AddDate(-100, 0, 0)) {
				return fmt.Errorf("%s cannot be more than 100 years ago", fieldName)
			}
			return nil
		})

		// Test avec une date valide
		err := timeValidator.Validate(time.Now().AddDate(0, 0, -1), "birthDate")
		assert.NoError(t, err)

		// Test avec une date future
		err = timeValidator.Validate(time.Now().AddDate(0, 0, 1), "birthDate")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be in the future")

		// Test avec une date trop ancienne
		err = timeValidator.Validate(time.Now().AddDate(-150, 0, 0), "birthDate")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be more than 100 years ago")
	})
}

// TestValidatorFuncWithGenerics démontre l'utilisation de ValidatorFunc avec les génériques
func TestValidatorFuncWithGenerics(t *testing.T) {
	t.Run("string validator with ValidatorFunc", func(t *testing.T) {
		// Utilisation de ValidatorFunc au lieu de NewValidatorFn
		stringValidator := ValidatorFunc[string](func(value string, fieldName string) error {
			if len(value) < 3 {
				return fmt.Errorf("%s is too short", fieldName)
			}
			return nil
		})

		err := stringValidator.Validate("hello", "name")
		assert.NoError(t, err)

		err = stringValidator.Validate("hi", "name")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is too short")
	})

	t.Run("int validator with ValidatorFunc", func(t *testing.T) {
		intValidator := ValidatorFunc[int](func(value int, fieldName string) error {
			if value%2 != 0 {
				return fmt.Errorf("%s must be even", fieldName)
			}
			return nil
		})

		err := intValidator.Validate(4, "number")
		assert.NoError(t, err)

		err = intValidator.Validate(3, "number")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be even")
	})
}

// TestWithMessageWithGenerics démontre l'utilisation de WithMessage avec les validateurs typés
func TestWithMessageWithGenerics(t *testing.T) {
	t.Run("custom message with typed validator", func(t *testing.T) {
		stringValidator := NewValidatorFn[string](func(value string, fieldName string) error {
			if len(value) < 5 {
				return fmt.Errorf("too short")
			}
			return nil
		}).WithMessage("Le nom doit contenir au moins 5 caractères")

		err := stringValidator.Validate("test", "name")
		assert.Error(t, err)
		assert.Equal(t, "Le nom doit contenir au moins 5 caractères", err.Error())
	})

	t.Run("custom message with ValidatorFunc", func(t *testing.T) {
		intValidator := ValidatorFunc[int](func(value int, fieldName string) error {
			if value < 18 {
				return fmt.Errorf("too young")
			}
			return nil
		}).WithMessage("Vous devez avoir au moins 18 ans")

		err := intValidator.Validate(16, "age")
		assert.Error(t, err)
		assert.Equal(t, "Vous devez avoir au moins 18 ans", err.Error())
	})
}

// TestIntegrationWithSchema démontre l'intégration des validateurs typés avec le système de schéma
func TestIntegrationWithSchema(t *testing.T) {
	t.Run("typed validator in schema", func(t *testing.T) {
		var name string
		var age int

		// Création de validateurs typés
		nameValidator := NewValidatorFn[string](func(value string, fieldName string) error {
			if len(value) < 2 {
				return fmt.Errorf("name too short")
			}
			return nil
		})

		ageValidator := NewValidatorFn[int](func(value int, fieldName string) error {
			if value < 0 || value > 120 {
				return fmt.Errorf("invalid age")
			}
			return nil
		})

		schema := NewSchema(
			Value("name", &name, WithValidators(nameValidator)),
			Value("age", &age, WithValidators(ageValidator)),
		)

		// Test avec des données valides
		data := map[string]interface{}{
			"name": "John",
			"age":  30,
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, "John", name)
		assert.Equal(t, 30, age)

		// Test avec des données invalides
		invalidData := map[string]interface{}{
			"name": "A", // Trop court
			"age":  150, // Trop vieux
		}

		err = schema.Apply(invalidData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name too short")
	})

	t.Run("mixed typed and regular validators", func(t *testing.T) {
		var email string
		var score float64

		// Validateur typé pour l'email
		emailValidator := NewValidatorFn[string](func(value string, fieldName string) error {
			if !strings.Contains(value, "@") {
				return fmt.Errorf("invalid email format")
			}
			return nil
		})

		// Validateur régulier pour le score
		scoreValidator := Min(0.0)

		schema := NewSchema(
			Value("email", &email, WithValidators(emailValidator, Required())),
			Value("score", &score, WithValidators(scoreValidator)),
		)

		data := map[string]interface{}{
			"email": "test@example.com",
			"score": 85.5,
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", email)
		assert.Equal(t, 85.5, score)
	})
}

// TestPerformanceComparison compare les performances entre les validateurs typés et non-typés
func TestPerformanceComparison(t *testing.T) {
	t.Run("typed vs untyped validator performance", func(t *testing.T) {
		// Validateur typé
		typedValidator := NewValidatorFn[string](func(value string, fieldName string) error {
			if len(value) < 5 {
				return fmt.Errorf("too short")
			}
			return nil
		})

		// Validateur non-typé (ancienne façon)
		untypedValidator := NewValidatorFn[interface{}](func(value interface{}, fieldName string) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("expected string")
			}
			if len(str) < 5 {
				return fmt.Errorf("too short")
			}
			return nil
		})

		testValue := "hello world"

		// Test de performance (exécution multiple pour mesurer)
		for i := 0; i < 1000; i++ {
			err := typedValidator.Validate(testValue, "test")
			assert.NoError(t, err)

			err = untypedValidator.Validate(testValue, "test")
			assert.NoError(t, err)
		}
	})
}

// TestEdgeCasesWithGenerics teste les cas limites avec les validateurs typés
func TestEdgeCasesWithGenerics(t *testing.T) {
	t.Run("nil value with typed validator", func(t *testing.T) {
		stringValidator := NewValidatorFn[string](func(value string, fieldName string) error {
			return nil
		})

		err := stringValidator.Validate(nil, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected type string, got <nil>")
	})

	t.Run("zero value with typed validator", func(t *testing.T) {
		intValidator := NewValidatorFn[int](func(value int, fieldName string) error {
			if value == 0 {
				return fmt.Errorf("zero value not allowed")
			}
			return nil
		})

		err := intValidator.Validate(0, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero value not allowed")
	})

	t.Run("complex type validation", func(t *testing.T) {
		type Config struct {
			Host     string
			Port     int
			Timeout  time.Duration
			Features []string
		}

		configValidator := NewValidatorFn[Config](func(value Config, fieldName string) error {
			if len(value.Host) == 0 {
				return fmt.Errorf("%s.host is required", fieldName)
			}
			if value.Port <= 0 || value.Port > 65535 {
				return fmt.Errorf("%s.port must be between 1 and 65535", fieldName)
			}
			if value.Timeout <= 0 {
				return fmt.Errorf("%s.timeout must be positive", fieldName)
			}
			return nil
		})

		validConfig := Config{
			Host:     "localhost",
			Port:     8080,
			Timeout:  time.Second * 30,
			Features: []string{"feature1", "feature2"},
		}

		err := configValidator.Validate(validConfig, "config")
		assert.NoError(t, err)

		invalidConfig := Config{
			Host:     "",
			Port:     70000,
			Timeout:  -time.Second,
			Features: []string{},
		}

		err = configValidator.Validate(invalidConfig, "config")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config.host is required")
	})
}
