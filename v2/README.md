# Poxxy v2

A type-safe, fluent validation library for Go with generics support.

## Features

- **Type-safe** validation with Go generics
- **Fluent API** for readable schema definitions
- **Built-in validators** for common use cases
- **Transformers** for data sanitization
- **i18n support** with French, Spanish, German translations
- **HTTP adapters** for request validation
- **Schema generation** for OpenAPI and JSON Schema

## Installation

```bash
go get github.com/arkan/poxxy/v2
```

## Quick Start

```go
package main

import (
    "fmt"
    "strings"

    "github.com/arkan/poxxy/v2"
)

func main() {
    var (
        name  string
        email string
        age   int
    )

    schema := poxxy.NewSchema(
        poxxy.Field("name", &name).
            Validate(poxxy.Required[string](), poxxy.MinLength[string](2)).
            Transform(poxxy.TrimSpace(), poxxy.Capitalize()).
            Describe("User's full name"),

        poxxy.Field("email", &email).
            Validate(poxxy.Required[string](), poxxy.Email()).
            Transform(poxxy.TrimSpace(), poxxy.ToLower()),

        poxxy.Field("age", &age).
            Validate(poxxy.Min(18), poxxy.Max(120)).
            Default(25),
    )

    err := schema.Parse(strings.NewReader(`{
        "name": "  john doe  ",
        "email": "  JOHN@EXAMPLE.COM  "
    }`))

    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Printf("Name: %s\n", name)   // Name: John doe
    fmt.Printf("Email: %s\n", email) // Email: john@example.com
    fmt.Printf("Age: %d\n", age)     // Age: 25
}
```

## Field Types

### Field - Required Values

```go
var name string
poxxy.Field("name", &name).
    Validate(poxxy.Required[string]()).
    Default("Anonymous")
```

### Pointer - Optional Values

```go
var nickname *string
poxxy.Pointer("nickname", &nickname).
    Validate(poxxy.MinLength[string](2))

// After parsing:
if nickname != nil {
    fmt.Println(*nickname)
}
```

### Slice - Arrays

```go
var tags []string
poxxy.Slice("tags", &tags).
    Validate(poxxy.MinItems[string](1), poxxy.MaxItems[string](10), poxxy.Unique[string]())
```

### Slice with Item Validation

```go
type Item struct {
    Name  string
    Price float64
}

var items []Item
poxxy.Slice("items", &items, func(item *Item, s *poxxy.Schema) {
    s.Add(poxxy.Field("name", &item.Name).Validate(poxxy.Required[string]()))
    s.Add(poxxy.Field("price", &item.Price).Validate(poxxy.Min(0.0)))
})
```

### Map - Key-Value Pairs

```go
var metadata map[string]string
poxxy.Map("metadata", &metadata)
```

### Struct - Nested Objects

```go
type Address struct {
    Street string
    City   string
    Zip    string
}

var address Address
poxxy.Struct("address", &address, func(s *poxxy.Schema) {
    s.Add(poxxy.Field("street", &address.Street).Validate(poxxy.Required[string]()))
    s.Add(poxxy.Field("city", &address.City).Validate(poxxy.Required[string]()))
    s.Add(poxxy.Field("zip", &address.Zip).Validate(poxxy.Pattern(`^\d{5}$`)))
})
```

### Convert - Type Conversion

```go
import "time"

var createdAt time.Time
poxxy.Convert("created_at", &createdAt, func(s string) (time.Time, error) {
    return time.Parse("2006-01-02", s)
}).Validate(poxxy.Required[time.Time]())
```

## Validators

### String Validators

```go
poxxy.Required[string]()           // Must not be empty
poxxy.MinLength[string](3)         // Minimum 3 characters
poxxy.MaxLength[string](50)        // Maximum 50 characters
poxxy.Pattern(`^[a-z]+$`)          // Regex pattern
poxxy.Email()                      // Valid email format
poxxy.URL()                        // Valid URL format
poxxy.UUID()                       // Valid UUID format
poxxy.In("active", "inactive")     // Must be one of values
poxxy.NotIn("banned", "deleted")   // Must not be one of values
```

### Numeric Validators

```go
poxxy.Min(18)                      // Minimum value
poxxy.Max(120)                     // Maximum value
poxxy.Between(18, 65)              // Between min and max
poxxy.Positive[int]()              // Must be > 0
poxxy.Negative[int]()              // Must be < 0
poxxy.NonZero[int]()               // Must not be 0
```

### Slice Validators

```go
poxxy.MinItems[string](1)          // Minimum 1 item
poxxy.MaxItems[string](10)         // Maximum 10 items
poxxy.Unique[string]()             // All items must be unique
poxxy.Each[string](poxxy.Email())  // Validate each item
```

### Custom Validators

```go
poxxy.Custom(func(value string, field string) error {
    if !strings.HasPrefix(value, "SK-") {
        return fmt.Errorf("%s must start with SK-", field)
    }
    return nil
})
```

### Conditional Validation

```go
var userType string
var companyName string

schema := poxxy.NewSchema(
    poxxy.Field("user_type", &userType).Validate(poxxy.In("personal", "business")),
    poxxy.Field("company_name", &companyName).Validate(
        poxxy.RequiredIf(func() bool { return userType == "business" }),
    ),
)
```

## Transformers

```go
poxxy.TrimSpace()                  // Remove whitespace
poxxy.ToLower()                    // Lowercase
poxxy.ToUpper()                    // Uppercase
poxxy.TitleCase()                  // Title Case
poxxy.Capitalize()                 // First letter uppercase
poxxy.SanitizeEmail()              // Normalize email
```

## Default Values

```go
// Default when field is absent
poxxy.Field("role", &role).Default("user")

// Default when field is absent OR empty
poxxy.Field("role", &role).DefaultOnEmpty("user")
```

## HTTP Integration

```go
import "github.com/arkan/poxxy/v2/adapters/http"

func handler(w http.ResponseWriter, r *http.Request) {
    var (
        name  string
        email string
    )

    schema := poxxy.NewSchema(
        poxxy.Field("name", &name).Validate(poxxy.Required[string]()),
        poxxy.Field("email", &email).Validate(poxxy.Required[string](), poxxy.Email()),
    )

    // Auto-detects content type (JSON, form, query params)
    if err := poxxyhttp.ApplyHTTPRequest(schema, r); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Use validated data
    fmt.Fprintf(w, "Hello %s!", name)
}
```

### Supported Content Types

- `application/json` - JSON body
- `application/x-www-form-urlencoded` - Form data
- `multipart/form-data` - Multipart forms
- Query parameters (fallback)

## Internationalization (i18n)

```go
import "github.com/arkan/poxxy/v2/i18n"

// Use French translations
schema := poxxy.NewSchema(
    poxxy.Field("email", &email).Validate(poxxy.Required[string]()),
).WithTranslator(i18n.French())

// Or Spanish, German...
schema.WithTranslator(i18n.Spanish())
schema.WithTranslator(i18n.German())

// Or detect from Accept-Language header
import "golang.org/x/text/language"
tag := language.French
schema.WithTranslator(i18n.ForLanguage(tag))
```

### Custom Messages

```go
catalog := i18n.New(nil).
    Set(i18n.MsgRequired, "Le champ {field} est obligatoire").
    Set(i18n.MsgEmail, "{field} doit être un email valide")

schema.WithTranslator(catalog)
```

## Error Handling

```go
err := schema.Parse(reader)
if err != nil {
    if verrs, ok := err.(*poxxy.ValidationErrors); ok {
        for _, e := range verrs.Errors {
            fmt.Printf("Field: %s, Rule: %s, Message: %s\n",
                e.Field, e.Rule, e.Message)
        }
    }
}
```

### Translating Errors

```go
verrs := err.(*poxxy.ValidationErrors)
translated := verrs.Translate(i18n.French())
fmt.Println(translated.Error())
```

## OpenAPI Schema Generation

```go
import "github.com/arkan/poxxy/v2/openapi"

schema := poxxy.NewSchema(
    poxxy.Field("name", &name).
        Validate(poxxy.Required[string](), poxxy.MinLength[string](2)).
        Describe("User's name"),
    poxxy.Field("age", &age).
        Validate(poxxy.Min(18), poxxy.Max(120)),
)

gen := openapi.NewGenerator()
openapiSchema := gen.FromSchema(schema)

jsonBytes, _ := openapiSchema.JSONIndent("", "  ")
fmt.Println(string(jsonBytes))
```

Output:
```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "User's name",
      "minLength": 2
    },
    "age": {
      "type": "integer",
      "format": "int32",
      "minimum": 18,
      "maximum": 120
    }
  },
  "required": ["name"]
}
```

## JSON Schema Generation

```go
import "github.com/arkan/poxxy/v2/jsonschema"

gen := jsonschema.NewGenerator(
    jsonschema.WithID("https://example.com/user.schema.json"),
    jsonschema.WithTitle("User"),
)
jsonSchema := gen.FromSchema(schema)

jsonBytes, _ := jsonSchema.JSONIndent("", "  ")
fmt.Println(string(jsonBytes))
```

## Templates (Reusable Schemas)

```go
// Define reusable field templates
emailField := poxxy.FieldTemplate[string]().
    Validate(poxxy.Required[string](), poxxy.Email()).
    Transform(poxxy.TrimSpace(), poxxy.ToLower())

urlField := poxxy.FieldTemplate[string]().
    Validate(poxxy.URL())

// Use in schemas
var email, website string
schema := poxxy.NewSchema(
    emailField.Build("email", &email),
    urlField.Build("website", &website),
)
```

### Built-in Presets

```go
poxxy.EmailTemplate()              // Required email with sanitization
poxxy.RequiredEmailTemplate()      // Same as EmailTemplate
poxxy.OptionalEmailTemplate()      // Optional email (pointer)
poxxy.URLTemplate()                // Required URL
poxxy.UUIDTemplate()               // Required UUID
poxxy.PositiveIntTemplate()        // Positive integer
poxxy.NonNegativeIntTemplate()     // >= 0 integer
```

## Deferred Validation (Async)

For validations requiring database or API calls:

```go
var email string
var checks poxxy.DeferredChecks

schema := poxxy.NewSchema(
    poxxy.Field("email", &email).Validate(
        poxxy.Required[string](),
        poxxy.Email(),
        poxxy.Deferred(&checks, func() error {
            // Check if email exists in database
            exists, err := db.EmailExists(email)
            if err != nil {
                return err
            }
            if exists {
                return fmt.Errorf("email already registered")
            }
            return nil
        }),
    ),
)

// Parse first (runs sync validators)
if err := schema.Parse(reader); err != nil {
    return err
}

// Then run deferred checks
if err := checks.Run(); err != nil {
    return err
}

// Or run in parallel
if err := checks.RunParallel(); err != nil {
    return err
}
```

## Strict Type Mode

Disable automatic type coercion:

```go
schema := poxxy.NewSchema(
    poxxy.Field("age", &age).StrictType(),  // Per-field
)

// Or for entire schema
schema.StrictTypes(true)
```

## Complete Example

```go
package main

import (
    "fmt"
    "net/http"
    "time"

    "github.com/arkan/poxxy/v2"
    poxxyhttp "github.com/arkan/poxxy/v2/adapters/http"
    "github.com/arkan/poxxy/v2/i18n"
)

type CreateUserRequest struct {
    Name      string
    Email     string
    Age       int
    Website   *string
    Tags      []string
    CreatedAt time.Time
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest

    schema := poxxy.NewSchema(
        poxxy.Field("name", &req.Name).
            Validate(poxxy.Required[string](), poxxy.MinLength[string](2), poxxy.MaxLength[string](100)).
            Transform(poxxy.TrimSpace(), poxxy.TitleCase()).
            Describe("User's full name"),

        poxxy.Field("email", &req.Email).
            Validate(poxxy.Required[string](), poxxy.Email()).
            Transform(poxxy.SanitizeEmail()),

        poxxy.Field("age", &req.Age).
            Validate(poxxy.Min(18), poxxy.Max(120)).
            Default(18),

        poxxy.Pointer("website", &req.Website).
            Validate(poxxy.URL()),

        poxxy.Slice("tags", &req.Tags).
            Validate(poxxy.MaxItems[string](5), poxxy.Unique[string]()),

        poxxy.Convert("created_at", &req.CreatedAt, func(s string) (time.Time, error) {
            return time.Parse(time.RFC3339, s)
        }).Default(time.Now()),
    ).WithTranslator(i18n.French())

    if err := poxxyhttp.ApplyHTTPRequest(schema, r); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Fprintf(w, "User created: %+v", req)
}

func main() {
    http.HandleFunc("/users", CreateUserHandler)
    http.ListenAndServe(":8080", nil)
}
```

## License

MIT License
