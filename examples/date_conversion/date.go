package main

import (
	"fmt"
	"time"

	"github.com/arkan/poxxy"
)

func main() {
	// Example demonstrating date conversion from string to time.Time
	var createdAt time.Time
	var updatedAt *time.Time
	var publishedAt time.Time

	schema := poxxy.NewSchema(
		// Convert string to time.Time
		poxxy.Convert("created_at", &createdAt, func(dateStr string) (time.Time, error) {
			return time.Parse("2006-01-02", dateStr)
		}, poxxy.WithValidators(poxxy.Required())),

		// Convert string to *time.Time (optional)
		poxxy.ConvertPointer("updated_at", &updatedAt, func(dateStr string) (time.Time, error) {
			return time.Parse("2006-01-02T15:04:05Z", dateStr)
		}),

		// Convert string to time.Time with default value
		poxxy.Convert("published_at", &publishedAt, func(dateStr string) (time.Time, error) {
			return time.Parse("2006-01-02", dateStr)
		},
			poxxy.WithDefault(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		),
	)

	data := map[string]interface{}{
		"created_at": "2024-01-15",
		"updated_at": "2024-01-15T10:30:00Z",
		// published_at will use default value
	}

	if err := schema.Apply(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("Created At: %v\n", createdAt)
	fmt.Printf("Updated At: %v\n", updatedAt)
	fmt.Printf("Published At: %v\n", publishedAt)

	// Example with different date formats
	var eventDate time.Time
	var eventTime time.Time

	dateSchema := poxxy.NewSchema(
		// Parse date in different format
		poxxy.Convert("event_date", &eventDate, func(dateStr string) (time.Time, error) {
			// Try multiple date formats
			formats := []string{
				"2006-01-02",
				"02/01/2006",
				"2006-01-02T15:04:05",
				"Jan 2, 2006",
			}

			for _, format := range formats {
				if t, err := time.Parse(format, dateStr); err == nil {
					return t, nil
				}
			}
			return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
		}),

		// Parse time with timezone
		poxxy.Convert("event_time", &eventTime, func(timeStr string) (time.Time, error) {
			return time.Parse("15:04:05", timeStr)
		}),
	)

	dateData := map[string]interface{}{
		"event_date": "15/01/2024",
		"event_time": "14:30:00",
	}

	if err := dateSchema.Apply(dateData); err != nil {
		fmt.Printf("Date validation failed: %v\n", err)
		return
	}

	fmt.Printf("Event Date: %v\n", eventDate)
	fmt.Printf("Event Time: %v\n", eventTime)
}
