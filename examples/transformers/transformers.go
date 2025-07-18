package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/arkan/poxxy"
)

func main() {
	// Custom transformer/converter example
	var timestamp time.Time
	var normalizedEmail string

	schema := poxxy.NewSchema(
		// Transform Unix timestamp to time.Time
		poxxy.Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (time.Time, error) {
			return time.Unix(unixTime, 0), nil
		}, poxxy.WithValidators(poxxy.Required())),

		// Normalize email to lowercase
		poxxy.Convert[string, string]("email", &normalizedEmail, func(email string) (string, error) {
			return strings.ToLower(strings.TrimSpace(email)), nil
		}, poxxy.WithValidators(poxxy.Required(), poxxy.Email())),
	)

	data := map[string]interface{}{
		"created_at": 1717689600,
		"email":      "John.Doe@example.com",
	}

	if err := schema.Apply(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("Timestamp: %v\n", timestamp)
	fmt.Printf("Normalized email: %s\n", normalizedEmail)
}
