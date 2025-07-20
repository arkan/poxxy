package poxxy

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertWithNewSignature(t *testing.T) {
	t.Run("convert returns pointer", func(t *testing.T) {
		var timestamp time.Time
		originalValue := timestamp

		schema := NewSchema(
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				t := time.Unix(unixTime, 0)
				return &t, nil
			}),
		)

		data := map[string]interface{}{
			"created_at": int64(1717689600),
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.NotEqual(t, originalValue, timestamp)
		assert.Equal(t, int64(1717689600), timestamp.Unix())
	})

	t.Run("convert sql.NullString", func(t *testing.T) {
		var createdAt sql.NullString

		schema := NewSchema(
			Convert[int64, sql.NullString]("created_at", &createdAt, func(unixTime int64) (*sql.NullString, error) {
				t := sql.NullString{String: time.Unix(unixTime, 0).Format(time.RFC3339), Valid: true}
				return &t, nil
			}),
		)

		data := map[string]interface{}{
			"created_at": int64(1717689600),
		}

		err := schema.Apply(data)
		assert.NoError(t, err)

		assert.Equal(t, "2024-06-06T18:00:00+02:00", createdAt.String)
		assert.Equal(t, true, createdAt.Valid)
	})

	t.Run("convert sql.NullString with required validator", func(t *testing.T) {
		var createdAt sql.NullString

		schema := NewSchema(
			Convert[int64, sql.NullString]("created_at", &createdAt, func(unixTime int64) (*sql.NullString, error) {
				t := sql.NullString{String: time.Unix(unixTime, 0).Format(time.RFC3339), Valid: true}
				return &t, nil
			}, WithValidators(Required())),
		)

		data := map[string]interface{}{
			"created_at": "",
		}

		err := schema.Apply(data)
		if assert.Error(t, err) {
			assert.Equal(t, "created_at: field is required", err.Error())
		}
	})

	t.Run("convert returns nil - should not mutate", func(t *testing.T) {
		var timestamp time.Time
		originalValue := timestamp

		schema := NewSchema(
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				// Return nil to indicate no conversion should happen
				return nil, nil
			}),
		)

		data := map[string]interface{}{
			"created_at": int64(1717689600),
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, originalValue, timestamp) // Should not be mutated
	})

	t.Run("convert returns error", func(t *testing.T) {
		var timestamp time.Time
		originalValue := timestamp

		schema := NewSchema(
			Convert[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				return nil, assert.AnError
			}),
		)

		data := map[string]interface{}{
			"created_at": int64(1717689600),
		}

		err := schema.Apply(data)
		assert.Error(t, err)
		assert.Equal(t, originalValue, timestamp) // Should not be mutated
	})
}

func TestConvertPointerWithNewSignature(t *testing.T) {
	t.Run("convert returns pointer", func(t *testing.T) {
		var timestamp *time.Time
		originalValue := timestamp

		schema := NewSchema(
			ConvertPointer[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				t := time.Unix(unixTime, 0)
				return &t, nil
			}),
		)

		data := map[string]interface{}{
			"created_at": int64(1717689600),
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.NotEqual(t, originalValue, timestamp)
		assert.NotNil(t, timestamp)
		assert.Equal(t, int64(1717689600), timestamp.Unix())
	})

	t.Run("convert returns nil - should not mutate", func(t *testing.T) {
		var timestamp *time.Time
		originalValue := timestamp

		schema := NewSchema(
			ConvertPointer[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				// Return nil to indicate no conversion should happen
				return nil, nil
			}),
		)

		data := map[string]interface{}{
			"created_at": int64(1717689600),
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, originalValue, timestamp) // Should not be mutated
	})

	t.Run("convert returns error", func(t *testing.T) {
		var timestamp *time.Time
		originalValue := timestamp

		schema := NewSchema(
			ConvertPointer[int64, time.Time]("created_at", &timestamp, func(unixTime int64) (*time.Time, error) {
				return nil, assert.AnError
			}),
		)

		data := map[string]interface{}{
			"created_at": int64(1717689600),
		}

		err := schema.Apply(data)
		assert.Error(t, err)
		assert.Equal(t, originalValue, timestamp) // Should not be mutated
	})
}

func TestConvertWithTransformers(t *testing.T) {
	t.Run("convert with transformers", func(t *testing.T) {
		var email string

		schema := NewSchema(
			Convert[string, string]("email", &email, func(emailStr string) (*string, error) {
				// Create a new string to avoid returning a pointer to the original
				result := emailStr
				return &result, nil
			}, WithTransformers(ToUpper())),
		)

		data := map[string]interface{}{
			"email": "john@example.com",
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, "JOHN@EXAMPLE.COM", email)
	})

	t.Run("convert returns nil with transformers", func(t *testing.T) {
		var email string
		originalValue := email

		schema := NewSchema(
			Convert[string, string]("email", &email, func(emailStr string) (*string, error) {
				return nil, nil // Return nil, should not apply transformers
			}, WithTransformers(ToUpper())),
		)

		data := map[string]interface{}{
			"email": "john@example.com",
		}

		err := schema.Apply(data)
		assert.NoError(t, err)
		assert.Equal(t, originalValue, email) // Should not be mutated
	})
}
