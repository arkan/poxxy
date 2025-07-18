package main

import (
	"fmt"

	"github.com/arkan/poxxy"
)

func main() {
	data := map[string]interface{}{
		"name":     "xxx",
		"location": "montains",
		"age":      10,
		"role":     nil,
	}

	var name string
	var location string
	var age int
	var role int

	schema := poxxy.NewSchema(
		poxxy.Value[string]("name", &name, poxxy.WithValidators(poxxy.Required())),
		poxxy.Value[string]("location", &location, poxxy.WithValidators(poxxy.Required())),
		poxxy.Value[int]("age", &age, poxxy.WithValidators(poxxy.Required())),
		poxxy.Value[int]("role", &role, poxxy.WithValidators(poxxy.Required())),
	)

	if err := schema.Apply(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Location: %s\n", location)
	fmt.Printf("Age: %d\n", age)
	fmt.Printf("Role: %d\n", role)
}
