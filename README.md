# Poxxy

A powerful Go library for data validation, transformation, and schema definition with support for default values and integrated transformers.

## Features

- **Type-safe schema definition** with generics
- **Built-in validators** for common validation rules
- **Default values** for fields when not provided
- **Integrated transformers** for data sanitization and formatting
- **Support for complex types** including structs, slices, arrays, and pointers
- **Comprehensive error handling** with detailed validation messages

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    "github.com/arkan/poxxy"
)

func main() {
    var name string
    var email string
    var age int
    var createdAt time.Time

    schema := poxxy.NewSchema(
        poxxy.Value("name", &name,
            poxxy.WithDefault("Anonymous"),
            poxxy.WithTransformers(
                poxxy.TrimSpace(),
                poxxy.Capitalize(),
            ),
            poxxy.WithValidators(poxxy.Required()),
        ),
        poxxy.Value("email", &email,
            poxxy.WithTransformers(poxxy.SanitizeEmail()),
            poxxy.WithValidators(poxxy.Required(), poxxy.Email()),
        ),
        poxxy.Value("age", &age,
            poxxy.WithDefault(25),
            poxxy.WithValidators(poxxy.Min(18), poxxy.Max(120)),
        ),
        poxxy.Convert("created_at", &createdAt, func(dateStr string) (time.Time, error) {
            return time.Parse("2006-01-02", dateStr)
        }, poxxy.WithDefault(time.Now())),
    )

    data := map[string]interface{}{
        "name":  "  john doe  ",
        "email": "  JOHN.DOE@EXAMPLE.COM  ",
        "created_at": "2024-01-15",
        // age will use default value
    }

    if err := schema.Apply(data); err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Printf("Name: '%s'\n", name)   // Output: Name: 'John doe'
    fmt.Printf("Email: '%s'\n", email) // Output: Email: 'john.doe@example.com'
    fmt.Printf("Age: %d\n", age)       // Output: Age: 25
    fmt.Printf("Created At: %v\n", createdAt) // Output: Created At: 2024-01-15 00:00:00 +0000 UTC
}
```

## Field Types

### Value Fields
Basic value fields for simple types like `string`, `int`, `bool`, etc.

```go
var name string
poxxy.Value("name", &name, opts...)
```

### Pointer Fields
Pointer fields for optional values or complex structs.

```go
var user *User
poxxy.Pointer("user", &user, opts...)
```

### Slice Fields
Slice fields for arrays of values or structs.

```go
var tags []string
poxxy.Slice("tags", &tags, opts...)
```

### Array Fields
Fixed-size array fields.

```go
var coords [2]float64
poxxy.Array("coords", &coords, opts...)
```

### Convert Fields
Fields that convert from one type to another (e.g., string to time.Time).

```go
var createdAt time.Time
poxxy.Convert("created_at", &createdAt, func(dateStr string) (time.Time, error) {
    return time.Parse("2006-01-02", dateStr)
}, opts...)

var updatedAt *time.Time
poxxy.ConvertPointer("updated_at", &updatedAt, func(dateStr string) (time.Time, error) {
    return time.Parse("2006-01-02T15:04:05Z", dateStr)
}, opts...)
```

## Options

### Default Values
Set default values for fields when they're not provided in the input data.

```go
poxxy.WithDefault("Anonymous")
poxxy.WithDefault(25)
poxxy.WithDefault(true)
```

### Transformers
Transform data before assignment and validation.

#### Built-in Transformers
- `ToUpper()` - Convert string to uppercase
- `ToLower()` - Convert string to lowercase
- `TrimSpace()` - Remove leading and trailing whitespace
- `TitleCase()` - Convert string to title case
- `Capitalize()` - Capitalize first letter
- `SanitizeEmail()` - Normalize email addresses

#### Custom Transformers
```go
poxxy.WithTransformers(
    poxxy.CustomTransformer(func(value string) (string, error) {
        return strings.ReplaceAll(value, "old", "new"), nil
    }),
)
```

### Validators
Apply validation rules to fields.

```go
poxxy.WithValidators(
    poxxy.Required(),
    poxxy.Email(),
    poxxy.Min(18),
    poxxy.Max(120),
    poxxy.MinLength(3),
    poxxy.MaxLength(50),
)
```

## Built-in Validators

- `Required()` - Field must be present and non-empty
- `Email()` - Valid email format
- `URL()` - Valid URL format
- `Min(value)` - Minimum numeric value
- `Max(value)` - Maximum numeric value
- `MinLength(length)` - Minimum string/slice length
- `MaxLength(length)` - Maximum string/slice length
- `In(values...)` - Value must be in the provided list
- `Each(validators...)` - Apply validators to each slice element
- `Unique()` - Slice elements must be unique
- `NotEmpty()` - Value must not be empty (non-zero)

## Examples

See the `examples/` directory for comprehensive examples:

- `examples/basic/` - Basic usage examples
- `examples/http_basic/` - HTTP request validation
- `examples/http_advanced/` - Advanced HTTP validation
- `examples/advanced_features/` - Default values and transformers
- `examples/date_conversion/` - Date and type conversion examples
- `examples/require/` - Required field validation
- `examples/debug/` - Debugging examples

### Date Conversion Example

```go
var createdAt time.Time
var updatedAt *time.Time

schema := poxxy.NewSchema(
    poxxy.Convert("created_at", &createdAt, func(dateStr string) (time.Time, error) {
        return time.Parse("2006-01-02", dateStr)
    }, poxxy.WithValidators(poxxy.Required())),

    poxxy.ConvertPointer("updated_at", &updatedAt, func(dateStr string) (time.Time, error) {
        return time.Parse("2006-01-02T15:04:05Z", dateStr)
    }),
)

data := map[string]interface{}{
    "created_at": "2024-01-15",
    "updated_at": "2024-01-15T10:30:00Z",
}
```

## Migration from Transform Fields

The `Transform` field type has been removed in favor of integrated transformers. Instead of:

```go
// Old way (removed)
poxxy.Transform[string, string]("email", &email, func(email string) (string, error) {
    return strings.ToLower(strings.TrimSpace(email)), nil
})
```

Use:

```go
// New way
poxxy.Value("email", &email,
    poxxy.WithTransformers(
        poxxy.SanitizeEmail(),
    ),
)
```

## License

MIT License - see LICENSE file for details.


