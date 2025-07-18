package main

import (
	"fmt"

	"github.com/arkan/poxxy"
)

func main() {
	var name string
	var age int
	var role string
	var salary int

	schema := poxxy.NewSchema(
		poxxy.Value("name", &name),
		poxxy.Value("age", &age),
		poxxy.Value("role", &role),
		poxxy.Value("salary", &salary),
	)

	schema.Apply(map[string]interface{}{
		"name":   "John",
		"age":    30,
		"role":   "ceo",
		"salary": 200000,
	})

	fmt.Printf("Name %s\n", name)
	fmt.Printf("Age %d\n", age)
	fmt.Printf("Role %s\n", role)
	fmt.Printf("Salary %d\n", salary)
}
