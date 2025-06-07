# Poxxy

[![Go Reference](https://pkg.go.dev/badge/github.com/arkan/poxxy.svg)](https://pkg.go.dev/github.com/arkan/poxxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/arkan/poxxy)](https://goreportcard.com/report/github.com/arkan/poxxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Poxxy is a powerful Go library that provides type-safe data mapping and validation for `map[string]interface{}` and HTTP form data. It combines the flexibility of data binding with comprehensive validation capabilities, making it ideal for handling user input in web applications, APIs, and configuration parsing. Unlike other validation libraries, Poxxy doesn't rely on struct tags - all validation rules are defined programmatically, giving you complete control and type safety.


## Key Features

- **Type-Safe Data Mapping** - Map dynamic data to strongly typed Go structs with compile-time safety
- **Comprehensive Validation** - Built-in validators for common use cases with custom validator support
- **Advanced Field Types** - Support for arrays, slices, pointers, nested structs, unions, and transformations
- **Schema-Based Definition** - Declarative API for defining data structure and validation rules
- **Conditional Validation** - Apply validators based on other field values
- **Rich Error Reporting** - Detailed validation errors with field-level granularity
- **Zero Allocation Friendly** - Efficient memory usage for high-performance applications

## Installation

```bash
go get github.com/arkan/poxxy
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/arkan/poxxy"
)

type User struct {
    Name  string
    Email string
    Age   int
}

func main() {
    // Input data (could come from HTTP forms, JSON, etc.)
    data := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    }

    var user User

    // Define schema with validation rules
    schema := poxxy.NewSchema(
        poxxy.Value("name", &user.Name, poxxy.WithValidators(
            poxxy.Required(),
            poxxy.MinLength(2),
        )),
        poxxy.Value("email", &user.Email, poxxy.WithValidators(
            poxxy.Required(),
            poxxy.Email(),
        )),
        poxxy.Value("age", &user.Age, poxxy.WithValidators(
            poxxy.Required(),
            poxxy.Min(18),
            poxxy.Max(120),
        )),
    )

    // Assign and validate
    if err := schema.Apply(data); err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Printf("User: %+v\n", user)
    // Output: User: {Name:John Doe Email:john@example.com Age:30}
}
```

## Advanced Usage

### Nested Structs

```go
type Address struct {
    Street string
    City   string
}

type User struct {
    Name    string
    Address Address
}

var user User

schema := poxxy.NewSchema(
    poxxy.Value("name", &user.Name, poxxy.WithValidators(poxxy.Required())),
    poxxy.Struct("address", &user.Address, func(s *poxxy.Schema, addr *Address) {
        poxxy.WithSchema(s, poxxy.Value("street", &addr.Street, poxxy.WithValidators(poxxy.Required())))
        poxxy.WithSchema(s, poxxy.Value("city", &addr.City, poxxy.WithValidators(poxxy.Required())))
    }),
)

data := map[string]interface{}{
    "name": "John",
    "address": map[string]interface{}{
        "street": "123 Main St",
        "city":   "Boston",
    },
}
```

### Optional Fields with Pointers

```go
type User struct {
    Name  string
    Email *string  // Optional
}

var user User

schema := poxxy.NewSchema(
    poxxy.Value("name", &user.Name, poxxy.WithValidators(poxxy.Required())),
    poxxy.Pointer("email", &user.Email, poxxy.WithValidators(poxxy.Email())),
)

// Email field can be omitted from data
data := map[string]interface{}{
    "name": "John",
    // email is optional
}
```

### Arrays and Slices

```go
var tags [3]string    // Fixed-size array
var scores []int      // Dynamic slice

schema := poxxy.NewSchema(
    // Fixed array
    poxxy.Array[string]("tags", &tags, poxxy.WithValidators(
        poxxy.Required(),
        poxxy.Each(poxxy.MinLength(1)),
        poxxy.Unique(),
    )),

    // Dynamic slice
    poxxy.Slice[int]("scores", &scores, poxxy.WithValidators(
        poxxy.MaxLength(10),
        poxxy.Each(poxxy.Min(0), poxxy.Max(100)),
    )),
)

data := map[string]interface{}{
    "tags":   [3]string{"work", "urgent", "important"},
    "scores": []int{95, 88, 92},
}
```

### Slice of Structs

```go
type House struct {
    Address string
    Price   int
}

type User struct {
    Name   string
    Houses []House
}

var user User

schema := poxxy.NewSchema(
    poxxy.Value("name", &user.Name, poxxy.WithValidators(poxxy.Required())),
    poxxy.SliceOf("houses", &user.Houses, func(s *poxxy.Schema, h *House) {
        poxxy.WithSchema(s, poxxy.Value("address", &h.Address, poxxy.WithValidators(
            poxxy.Required(),
            poxxy.MinLength(5),
        )))
        poxxy.WithSchema(s, poxxy.Value("price", &h.Price, poxxy.WithValidators(
            poxxy.Min(0),
            poxxy.Max(1000000),
        )))
    }, poxxy.WithValidators(
        poxxy.MinLength(1),
        poxxy.MaxLength(10),
    )),
)

data := map[string]interface{}{
    "name": "John",
    "houses": []map[string]interface{}{
        {
            "address": "123 Main St",
            "price":   100000,
        },
        {
            "address": "456 Elm St",
            "price":   200000,
        },
    },
}
```

### Type Transformations

```go
var timestamp time.Time
var normalizedEmail string

schema := poxxy.NewSchema(
    // Transform Unix timestamp to time.Time
    poxxy.Transform[int64, time.Time]("created_at", &timestamp,
        func(unixTime int64) (time.Time, error) {
            return time.Unix(unixTime, 0), nil
        }, poxxy.WithValidators(poxxy.Required())),

    // Normalize email to lowercase
    poxxy.Transform[string, string]("email", &normalizedEmail,
        func(email string) (string, error) {
            return strings.ToLower(strings.TrimSpace(email)), nil
        }, poxxy.WithValidators(poxxy.Required(), poxxy.Email())),
)

data := map[string]interface{}{
    "created_at": int64(1717689600),
    "email":      "John.Doe@EXAMPLE.COM",
}
```

### Union/Polymorphic Types

```go
type Document interface {
    GetType() string
}

type TextDocument struct {
    Content   string
    WordCount int
}

func (t TextDocument) GetType() string { return "text" }

type ImageDocument struct {
    URL    string
    Width  int
    Height int
}

func (i ImageDocument) GetType() string { return "image" }

var document Document

schema := poxxy.NewSchema(
    poxxy.Union("document", &document, func(data map[string]interface{}) (interface{}, error) {
        docType, ok := data["type"].(string)
        if !ok {
            return nil, fmt.Errorf("missing document type")
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
    }),
)
```

## Built-in Validators

### Basic Validators
- `Required()` - Field must be present in input data
- `NonZero()` - Value must not be zero value
- `Email()` - Valid email format
- `URL()` - Valid URL format

### Numeric Validators
- `Min(value)` - Minimum value for numbers
- `Max(value)` - Maximum value for numbers

### String & Collection Validators
- `MinLength(n)` - Minimum length for strings/slices/arrays
- `MaxLength(n)` - Maximum length for strings/slices/arrays
- `In(values...)` - Value must be one of the specified values

### Collection Validators
- `Each(validators...)` - Apply validators to each element
- `Unique()` - All elements must be unique
- `UniqueBy(keyExtractor)` - Elements must be unique by extracted key

### Custom Error Messages

```go
poxxy.Value("age", &age, poxxy.WithValidators(
    poxxy.Min(18).WithMessage("You must be at least 18 years old"),
    poxxy.Max(120).WithMessage("Age cannot exceed 120 years"),
))
```

## Supported Field Types

### Basic Types
- `string`, `int`, `int64`, `float64`, `bool`
- Pointers to basic types (for optional fields)

### Complex Types
- `[]T` - Slices of any type
- `[N]T` - Fixed-size arrays
- Nested structs with validation callbacks
- `map[K]V` - Maps with key/value validation
- Interface types with union resolvers

### Advanced Types
- Type transformations with `Transform[From, To]()`
- Conditional validation based on other fields
- Custom validators and field types

## Error Handling

Poxxy provides detailed error information for validation failures:

```go
if err := schema.Apply(data); err != nil {
    if validationErrors, ok := err.(poxxy.Errors); ok {
        for _, fieldError := range validationErrors {
            fmt.Printf("Field '%s': %v\n", fieldError.Field, fieldError.Error)
        }
    }
}
```

## Roadmap

### Planned Features

- [ ] **Conditional Validation** - Full implementation of field-dependent validation
- [ ] **Localization Support** - i18n error messages
- [ ] **JSON Schema Export** - Generate JSON Schema from Poxxy schemas
- [ ] **HTTP Integration** - Direct HTTP request parsing utilities
- [ ] **Performance Optimizations** - Reflection caching and code generation
- [ ] **Advanced Validators** - Credit card, phone number, date range validators

### Future Considerations

- Real-time validation for streaming data
- Plugin system for custom validators

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Inspiration

Poxxy draws inspiration from excellent Go libraries:

- [go-playground/form](https://github.com/go-playground/form) - For flexible form decoding patterns
- [go-playground/validator](https://github.com/go-playground/validator) - For comprehensive validation design
- [invopop/validation](https://github.com/invopop/validation) - For fluent validation API concepts


