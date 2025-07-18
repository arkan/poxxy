package main

import (
	"fmt"
	"strings"

	"github.com/arkan/poxxy"
)

func main() {
	// Example demonstrating new features:
	// 1. Default values
	// 2. Integrated transformers
	// 3. No more Transform field needed

	var name string
	var email string
	var age int
	var isActive bool
	var tags []string

	schema := poxxy.NewSchema(
		// String with default value and transformers
		poxxy.Value("name", &name,
			poxxy.WithDefault("Anonymous"),
			poxxy.WithTransformers(
				poxxy.TrimSpace(),
				poxxy.Capitalize(),
			),
			poxxy.WithValidators(poxxy.Required()),
		),

		// Email with sanitization transformer
		poxxy.Value("email", &email,
			poxxy.WithTransformers(
				poxxy.SanitizeEmail(),
			),
			poxxy.WithValidators(poxxy.Required(), poxxy.Email()),
		),

		// Age with default value
		poxxy.Value("age", &age,
			poxxy.WithDefault(25),
			poxxy.WithValidators(poxxy.Min(18), poxxy.Max(120)),
		),

		// Boolean with default value
		poxxy.Value("is_active", &isActive,
			poxxy.WithDefault(true),
		),

		// Slice with default value and transformers
		poxxy.Slice("tags", &tags,
			poxxy.WithDefault([]string{"general"}),
			poxxy.WithTransformers(
				poxxy.CustomTransformer(func(tags []string) ([]string, error) {
					// Remove duplicates and sort
					seen := make(map[string]bool)
					result := []string{}
					for _, tag := range tags {
						normalized := strings.ToLower(strings.TrimSpace(tag))
						if !seen[normalized] {
							seen[normalized] = true
							result = append(result, normalized)
						}
					}
					return result, nil
				}),
			),
		),
	)

	// Test with partial data
	data := map[string]interface{}{
		"name":  "  john doe  ",
		"email": "  john.doe@example.com  ",
		"age":   30, // Provide a valid age
		"tags":  []string{"TECH", "tech", "  programming  "},
		// is_active will use default value
	}

	if err := schema.Apply(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("Name: '%s'\n", name)
	fmt.Printf("Email: '%s'\n", email)
	fmt.Printf("Age: %d\n", age)
	fmt.Printf("Is Active: %t\n", isActive)
	fmt.Printf("Tags: %v\n", tags)

	// Test with pointer fields
	var title *string
	var score *float64

	pointerSchema := poxxy.NewSchema(
		// Pointer with default value and transformers
		poxxy.Pointer("title", &title,
			poxxy.WithDefault("Untitled"),
			poxxy.WithTransformers(
				poxxy.TrimSpace(),
				poxxy.TitleCase(),
			),
		),

		// Pointer with default value
		poxxy.Pointer("score", &score,
			poxxy.WithDefault(0.0),
		),
	)

	pointerData := map[string]interface{}{
		"title": "  hello world  ",
		// score will use default value
	}

	if err := pointerSchema.Apply(pointerData); err != nil {
		fmt.Printf("Pointer validation failed: %v\n", err)
		return
	}

	fmt.Printf("Title: '%s'\n", *title)
	if score != nil {
		fmt.Printf("Score: %f\n", *score)
	} else {
		fmt.Printf("Score: nil\n")
	}
}
