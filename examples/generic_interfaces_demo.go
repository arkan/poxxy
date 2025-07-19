package main

import (
	"fmt"

	"github.com/arkan/poxxy"
)

// This example demonstrates the improvement brought by generic interfaces
// Before: fragile type switching in validators.go (lines 60-90)
// After: robust generic interfaces with ValidatorsAppender and DefaultValueSetter

func main() {
	fmt.Println("=== Demonstration of Improved Generic Interfaces ===")

	// 1. Different field types with validators - all use the same interface
	var name string
	var age int
	var scores []float64
	var settings map[string]string

	schema := poxxy.NewSchema(
		// ValueField
		poxxy.Value("name", &name,
			poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(2))),

		// ValueField with different type
		poxxy.Value("age", &age,
			poxxy.WithValidators(poxxy.Required(), poxxy.Min(18), poxxy.Max(120))),

		// SliceField
		poxxy.Slice("scores", &scores,
			poxxy.WithValidators(poxxy.Required(), poxxy.Each(poxxy.Min(0.0), poxxy.Max(100.0)))),

		// MapField
		poxxy.Map("settings", &settings,
			poxxy.WithValidators(poxxy.Required())),
	)

	data := map[string]interface{}{
		"name":   "Alice",
		"age":    25,
		"scores": []float64{85.5, 92.0, 78.5},
		"settings": map[string]interface{}{
			"theme": "dark",
			"lang":  "en",
		},
	}

	if err := schema.Apply(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ All fields validated successfully!\n")
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Age: %d\n", age)
	fmt.Printf("Scores: %v\n", scores)
	fmt.Printf("Settings: %v\n\n", settings)

	// 2. Demonstration of default values with generic interface
	var title string
	var count int
	var active bool

	defaultSchema := poxxy.NewSchema(
		poxxy.Value("title", &title, poxxy.WithDefault("Untitled")),
		poxxy.Value("count", &count, poxxy.WithDefault(0)),
		poxxy.Value("active", &active, poxxy.WithDefault(true)),
	)

	// Empty data - default values will be used
	emptyData := map[string]interface{}{}

	if err := defaultSchema.Apply(emptyData); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Default values applied automatically!\n")
	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Count: %d\n", count)
	fmt.Printf("Active: %t\n\n", active)

	// 3. Demonstration of extensibility
	fmt.Println("To add a new field type, simply implement:")
	fmt.Println("- ValidatorsAppender for validators")
	fmt.Println("- DefaultValueSetter[T] for default values")
	fmt.Println("- The Field interface for complete integration")

	fmt.Println("=== Advantages of the New Approach ===")
	fmt.Println("✅ No more fragile type switching")
	fmt.Println("✅ Easy extensibility - just implement the interfaces")
	fmt.Println("✅ Reduced maintenance - no need to add new cases")
	fmt.Println("✅ Improved type safety with generics")
	fmt.Println("✅ More readable and maintainable code")
}
