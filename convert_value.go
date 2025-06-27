package poxxy

import (
	"database/sql"
	"fmt"

	"github.com/arkan/go-convert"
)

// convertValue converts interface{} to type T
func convertValue[T any](value interface{}) (T, error) {
	var zero T

	// Direct type assertion first
	if v, ok := value.(T); ok {
		return v, nil
	}

	// Handle sql.Null types (e.g. sql.NullString, sql.NullInt64)
	if v, ok := any(&zero).(sql.Scanner); ok {
		err := v.Scan(value)
		if err != nil {
			return zero, err
		}

		return zero, nil
	}

	// Convert using go-convert
	err := convert.Convert(value, &zero)
	if err == nil {
		return zero, nil
	}

	return zero, fmt.Errorf("cannot convert %T to %T", value, zero)
}
