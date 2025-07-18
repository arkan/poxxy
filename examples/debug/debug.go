package main

import (
	"fmt"

	"github.com/arkan/poxxy"
)

func main() {
	// Simple test with just email
	var email string

	schema := poxxy.NewSchema(
		poxxy.Value("email", &email,
			poxxy.WithTransformers(
				poxxy.SanitizeEmail(),
			),
			poxxy.WithValidators(poxxy.Required(), poxxy.Email()),
		),
	)

	data := map[string]interface{}{
		"email": "  john.doe@example.com  ",
	}

	if err := schema.Apply(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("Email: '%s'\n", email)

	// Test default value
	var name string

	nameSchema := poxxy.NewSchema(
		poxxy.Value("name", &name,
			poxxy.WithDefault("Anonymous"),
			poxxy.WithTransformers(
				poxxy.TrimSpace(),
				poxxy.Capitalize(),
			),
		),
	)

	// No name in data, should use default
	nameData := map[string]interface{}{}

	if err := nameSchema.Apply(nameData); err != nil {
		fmt.Printf("Name validation failed: %v\n", err)
		return
	}

	fmt.Printf("Name: '%s'\n", name)
}
