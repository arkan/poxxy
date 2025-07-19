# Poxxy

A powerful Go library for data validation, transformation, and schema definition with support for default values and integrated transformers.

## Features

- **Type-safe schema definition** with generics
- **Built-in validators** for common validation rules
- **Default values** for fields when not provided
- **Integrated transformers** for data sanitization and formatting
- **Support for complex types** including structs, slices, arrays, pointers, maps, and unions
- **HTTP request validation** with automatic content-type detection
- **Comprehensive error handling** with detailed validation messages
- **Two-phase validation** (assignment then validation) for cross-field validation

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

### Map Fields
Map fields for key-value pairs with validation.

```go
var settings map[string]string
poxxy.Map("settings", &settings, opts...)

// With default values
defaultSettings := map[string]string{
    "theme": "dark",
    "lang":  "en",
}
poxxy.Map("settings", &settings, poxxy.WithDefault(defaultSettings), opts...)
```

### NestedMap Fields
Nested map fields with validation for each key-value pair.

```go
var userSettings map[string]string
poxxy.NestedMap("settings", &userSettings, func(s *poxxy.Schema, key string, value *string) {
    // Validate each key-value pair
    poxxy.WithSchema(s, poxxy.ValueWithoutAssign("key", poxxy.WithValidators(poxxy.Required())))
    poxxy.WithSchema(s, poxxy.ValueWithoutAssign("value", poxxy.WithValidators(poxxy.Required())))
}, opts...)
```

### HTTPMap Fields
HTTPMap fields for handling form-encoded map data from HTTP requests. Each map value is a struct with its own schema.

```go
type Attachment struct {
    URL      string `json:"url"`
    Filename string `json:"filename"`
    Size     int    `json:"size"`
}

var attachments map[string]Attachment
defaultAttachments := map[string]Attachment{
    "default": {
        URL:      "https://example.com/default.pdf",
        Filename: "default.pdf",
        Size:     1024,
    },
}

poxxy.HTTPMap("attachments", &attachments, func(s *poxxy.Schema, a *Attachment) {
    poxxy.WithSchema(s, poxxy.Value("url", &a.URL, poxxy.WithValidators(poxxy.Required())))
    poxxy.WithSchema(s, poxxy.Value("filename", &a.Filename, poxxy.WithValidators(poxxy.Required())))
    poxxy.WithSchema(s, poxxy.Value("size", &a.Size, poxxy.WithDefault(0)))
}, poxxy.WithDefault(defaultAttachments))
```

This handles form data like:
```
attachments[0][url]=https://example.com/doc1.pdf&attachments[0][filename]=doc1.pdf&attachments[0][size]=2048
```

### Struct Fields
Struct fields for nested object validation with sub-schemas.

```go
type Address struct {
    Street string
    City   string
}

var address Address
poxxy.Struct("address", &address, poxxy.WithSubSchema(func(s *poxxy.Schema, addr *Address) {
    poxxy.WithSchema(s, poxxy.Value("street", &addr.Street, poxxy.WithValidators(poxxy.Required())))
    poxxy.WithSchema(s, poxxy.Value("city", &addr.City, poxxy.WithValidators(poxxy.Required())))
}), opts...)
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

### Union Fields
Union/polymorphic fields for handling different types based on a discriminator.

```go
type Document interface {
    GetType() string
}

type TextDocument struct {
    Content   string
    WordCount int
}

type ImageDocument struct {
    URL    string
    Width  int
    Height int
}

var document Document
poxxy.Union("document", &document, func(data map[string]interface{}) (interface{}, error) {
    docType, ok := data["type"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid document type")
    }

    switch docType {
    case "text":
        var doc TextDocument
        subSchema := poxxy.NewSchema(
            poxxy.Value("content", &doc.Content, poxxy.WithValidators(poxxy.Required())),
            poxxy.Value("word_count", &doc.WordCount, poxxy.WithValidators(poxxy.Min(0))),
        )
        if err := subSchema.Apply(data); err != nil {
            return nil, err
        }
        return doc, nil
    case "image":
        var doc ImageDocument
        subSchema := poxxy.NewSchema(
            poxxy.Value("url", &doc.URL, poxxy.WithValidators(poxxy.Required(), poxxy.URL())),
            poxxy.Value("width", &doc.Width, poxxy.WithValidators(poxxy.Min(1))),
            poxxy.Value("height", &doc.Height, poxxy.WithValidators(poxxy.Min(1))),
        )
        if err := subSchema.Apply(data); err != nil {
            return nil, err
        }
        return doc, nil
    default:
        return nil, fmt.Errorf("unknown document type: %s", docType)
    }
})
```

### ValueWithoutAssign Fields
Fields that validate values without assigning them to variables (useful in map validation).

```go
poxxy.ValueWithoutAssign("key", poxxy.WithValidators(poxxy.Required()))
```

## Options

### Default Values
Set default values for fields when they're not provided in the input data.

```go
poxxy.WithDefault("Anonymous")
poxxy.WithDefault(25)
poxxy.WithDefault(true)
poxxy.WithDefault(map[string]string{"theme": "dark"})
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

### Descriptions
Add descriptions to fields for better error messages and documentation.

```go
poxxy.WithDescription("User's full name")
```

### Sub-schema Options
Configure sub-schemas for complex nested structures.

```go
poxxy.WithSubSchema(func(s *poxxy.Schema, user *User) {
    // Configure sub-schema fields
})

poxxy.WithSubSchemaMap(func(s *poxxy.Schema, key string, value *string) {
    // Configure map value validation
})
```

## Built-in Validators

### Basic Validators
- `Required()` - Field must be present and non-empty
- `NotEmpty()` - Value must not be empty (non-zero)
- `Email()` - Valid email format
- `URL()` - Valid URL format (http/https only)

### Numeric Validators
- `Min(value)` - Minimum numeric value
- `Max(value)` - Maximum numeric value

### String and Collection Validators
- `MinLength(length)` - Minimum string/slice length
- `MaxLength(length)` - Maximum string/slice length
- `In(values...)` - Value must be in the provided list
- `Unique()` - Slice/array/map elements must be unique
- `UniqueBy(keyExtractor)` - Elements must be unique by extracted key

### Collection Validators
- `Each(validators...)` - Apply validators to each slice/array element
- `WithMapKeys(keys...)` - Map must contain specified keys

### Custom Validators
```go
poxxy.ValidatorFunc(func(value string, fieldName string) error {
    if !strings.Contains(value, "@") {
        return fmt.Errorf("must contain @ symbol")
    }
    return nil
})
```

### Validator Messages
Customize error messages for validators.

```go
poxxy.Required().WithMessage("This field is mandatory")
poxxy.Min(18).WithMessage("Must be at least 18 years old")
```

## Schema Options

### Skip Validators
Skip validation phase (useful for data transformation only).

```go
schema.Apply(data, poxxy.WithSkipValidators(true))
```

## HTTP Integration

### ApplyHTTPRequest
Validate HTTP requests with automatic content-type detection.

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    var user User
    schema := poxxy.NewSchema(
        poxxy.Value("name", &user.Name, poxxy.WithValidators(poxxy.Required())),
        poxxy.Value("email", &user.Email, poxxy.WithValidators(poxxy.Required(), poxxy.Email())),
    )

    if err := schema.ApplyHTTPRequest(r); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Process validated user data
}
```

### ApplyJSON
Validate JSON data directly.

```go
jsonData := `{"name": "John", "email": "john@example.com"}`
if err := schema.ApplyJSON([]byte(jsonData)); err != nil {
    // Handle validation error
}
```

### Supported Content Types
- `application/json` - JSON request body
- `application/x-www-form-urlencoded` - Form data
- No content type - Query parameters

## Error Handling

### Error Types
- `FieldError` - Individual field validation error
- `Errors` - Collection of field errors

### Error Information
Each error includes:
- Field name
- Error description
- Field description (if provided)

```go
if err := schema.Apply(data); err != nil {
    if errors, ok := err.(poxxy.Errors); ok {
        for _, fieldError := range errors {
            fmt.Printf("Field '%s': %v\n", fieldError.Field, fieldError.Error)
        }
    }
}
```

## Advanced Examples

### Complex Nested Structure
```go
type User struct {
    Name     string
    Email    *string
    Address  *Address
    Settings map[string]interface{}
    Tags     []string
    Scores   [5]int
}

var user User
schema := poxxy.NewSchema(
    poxxy.Struct("user", &user, poxxy.WithSubSchema(func(s *poxxy.Schema, u *User) {
        poxxy.WithSchema(s, poxxy.Value("name", &u.Name, poxxy.WithValidators(poxxy.Required())))
        poxxy.WithSchema(s, poxxy.Pointer("email", &u.Email, poxxy.WithValidators(poxxy.Email())))
        poxxy.WithSchema(s, poxxy.Pointer("address", &u.Address, poxxy.WithSubSchema(func(ss *poxxy.Schema, addr *Address) {
            poxxy.WithSchema(ss, poxxy.Value("street", &addr.Street, poxxy.WithValidators(poxxy.Required())))
            poxxy.WithSchema(ss, poxxy.Value("city", &addr.City, poxxy.WithValidators(poxxy.Required())))
        })))
        poxxy.WithSchema(s, poxxy.Map("settings", &u.Settings))
        poxxy.WithSchema(s, poxxy.Slice("tags", &u.Tags, poxxy.WithValidators(poxxy.Unique())))
        poxxy.WithSchema(s, poxxy.Array("scores", &u.Scores, poxxy.WithValidators(poxxy.Each(poxxy.Min(0), poxxy.Max(100)))))
    })),
)
```

### Date and Type Conversion
```go
var createdAt time.Time
var updatedAt *time.Time
var unixTimestamp int64

schema := poxxy.NewSchema(
    poxxy.Convert("created_at", &createdAt, func(dateStr string) (time.Time, error) {
        return time.Parse("2006-01-02", dateStr)
    }, poxxy.WithValidators(poxxy.Required())),

    poxxy.ConvertPointer("updated_at", &updatedAt, func(dateStr string) (time.Time, error) {
        return time.Parse("2006-01-02T15:04:05Z", dateStr)
    }),

    poxxy.Convert("timestamp", &unixTimestamp, func(unixTime int64) (int64, error) {
        return unixTime, nil
    }, poxxy.WithTransformers(poxxy.CustomTransformer(func(timestamp int64) (int64, error) {
        // Ensure timestamp is not in the future
        if timestamp > time.Now().Unix() {
            return 0, fmt.Errorf("timestamp cannot be in the future")
        }
        return timestamp, nil
    }))),
)
```

## Examples Directory

See the `examples/` directory for comprehensive examples:

- `examples/basic/` - Basic usage examples
- `examples/http_basic/` - HTTP request validation
- `examples/http_advanced/` - Advanced HTTP validation
- `examples/advanced_features/` - Default values and transformers
- `examples/date_conversion/` - Date and type conversion examples
- `examples/require/` - Required field validation
- `examples/debug/` - Debugging examples
- `examples/transformers/` - Custom transformers and converters
- `examples/dsl_v2/` - DSL v2 examples

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

## Performance Considerations

- **Two-phase validation**: Assignment happens before validation to ensure all fields are populated for cross-field validation
- **Lazy evaluation**: Validators are only applied when needed
- **Type safety**: Generics ensure compile-time type safety throughout the validation pipeline
- **Memory efficient**: Minimal allocations during validation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

MIT License - see LICENSE file for details.


