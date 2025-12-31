# Poxxy V2 - Specification

## Overview

Poxxy V2 is a complete rewrite of the data validation and assignment library for Go. It provides a declarative, type-safe DSL using Go generics without relying on struct tags.

**Key principles:**
- No struct tags
- Generics-first for maximum type safety
- Compile-time validation of validator/transformer usage
- Minimal reflection
- Fluent API with reduced verbosity

---

## Breaking Changes

V2 is a **breaking change** from V1. It will use a new module path:
```
github.com/arkan/poxxy/v2
```

---

## API Design

### Field Declaration - Fluent API

```go
// V2 Fluent API
poxxy.Field("email", &email).
    Default("").
    Transform(poxxy.TrimSpace(), poxxy.SanitizeEmail()).
    Validate(poxxy.Required(), poxxy.Email()).
    Describe("User email")
```

### Schema Creation

```go
schema := poxxy.NewSchema(
    poxxy.Field("name", &user.Name).Validate(poxxy.Required()),
    poxxy.Field("email", &user.Email).
        Transform(poxxy.TrimSpace(), poxxy.ToLower()).
        Validate(poxxy.Required(), poxxy.Email()),
    poxxy.Field("age", &user.Age).Validate(poxxy.Min(0), poxxy.Max(150)),
)

err := schema.Parse(jsonReader)
```

### Main Method: Parse()

The main method is `Parse()` (not `Apply()`) to suggest decoding + validation:

```go
func (s *Schema) Parse(r io.Reader) error
func (s *Schema) ParseDecoder(d Decoder) error
```

---

## Field Types

### Kept from V1:
| Type | Generic Signature | Description |
|------|-------------------|-------------|
| **Value** | `Field[T]` | Basic scalar values |
| **Pointer** | `Pointer[T]` | Optional/nullable values |
| **Slice** | `Slice[T]` | Collections |
| **Map** | `Map[K, V]` | Key-value pairs |
| **Struct** | `Struct[T]` | Nested objects with sub-schema |
| **Convert** | `Convert[From, To]` | Type conversion at assignment |

### Removed from V1:
- `ArrayField` - Use `Slice` + length validation
- `NestedMapField` - Use `Map` + `Each()` validator
- `UnionField` - User handles polymorphism manually
- `HTTPMapField` - Use adapter pattern instead

---

## Templates - Reusable Schemas

Templates are **strictly immutable** - no override allowed:

```go
// Define reusable templates
var EmailField = poxxy.Template[string]().
    Transform(poxxy.TrimSpace(), poxxy.ToLower()).
    Validate(poxxy.Required(), poxxy.Email())

var AddressSchema = poxxy.StructTemplate[Address](func(s *poxxy.SchemaBuilder[Address]) {
    s.Field("street", &s.Target.Street).Validate(poxxy.Required())
    s.Field("city", &s.Target.City).Validate(poxxy.Required())
    s.Field("zip", &s.Target.Zip).Validate(poxxy.Required(), poxxy.Pattern(`^\d{5}$`))
})

// Use templates - NO override, create new template if different behavior needed
schema := poxxy.NewSchema(
    EmailField.Bind("email", &user.Email),
    EmailField.Bind("backup_email", &user.BackupEmail), // Same validators
    AddressSchema.Bind("address", &user.Address),
)
```

---

## Validators

### Type-Safe Validators

Validators are typed with generics. Compile-time errors if wrong type:

```go
// Typed validators
func Required[T any]() Validator[T]
func Email() Validator[string]
func URL() Validator[string]
func Min[T constraints.Ordered](value T) Validator[T]
func Max[T constraints.Ordered](value T) Validator[T]
func MinLength[T ~string | ~[]byte](n int) Validator[T]
func MaxLength[T ~string | ~[]byte](n int) Validator[T]
func In[T comparable](values ...T) Validator[T]
func Each[T any](validators ...Validator[T]) Validator[[]T]
func Unique[T comparable]() Validator[[]T]
func UniqueBy[T any, K comparable](keyFn func(T) K) Validator[[]T]
func Pattern(regex string) Validator[string]
```

### Standalone Functions with Type Inference

Validators are standalone functions that infer type from `FieldBuilder[T].Validate()`:

```go
// Type inferred from field
poxxy.Field("email", &email).Validate(poxxy.MinLength(3)) // MinLength[string]
poxxy.Field("users", &users).Validate(poxxy.MinItems(1))  // MinItems[[]User]
```

### Cross-Field Validators

For conditional validation based on other fields:

```go
poxxy.Field("billing_address", &billing).
    Validate(poxxy.RequiredIf(func(s *Schema) bool {
        return s.Get("use_separate_billing").(bool)
    }))
```

### Custom Validators

Helper function for inline custom validators:

```go
poxxy.ValidatorFunc(func(v string) error {
    if !isValidDomain(v) {
        return errors.New("invalid domain")
    }
    return nil
})
```

### Async Validators (Callback Pattern)

For validations requiring external calls (DB, API):

```go
// Validators can return deferred checks
type DeferredValidator[T any] interface {
    Validate(value T) (func() error, error)
}

// Usage
poxxy.DeferredValidator(func(email string) func() error {
    return func() error {
        exists, err := db.EmailExists(email)
        if err != nil { return err }
        if exists { return errors.New("email already registered") }
        return nil
    }
})
```

---

## Transformers

### Type-Safe Transformers

Transformers are strictly typed - compile-time error if wrong type:

```go
func TrimSpace() Transformer[string]
func ToLower() Transformer[string]
func ToUpper() Transformer[string]
func TitleCase() Transformer[string]
func Capitalize() Transformer[string]
func SanitizeEmail() Transformer[string]
func Abs[T constraints.Signed | constraints.Float]() Transformer[T]
```

### Custom Transformers

```go
poxxy.TransformerFunc(func(s string) (string, error) {
    return strings.ReplaceAll(s, " ", "-"), nil
})
```

---

## Error Handling

### Structured Errors

```go
type ValidationError struct {
    Field   string   // Flat field name: "users[2].email"
    Value   any      // The invalid value
    Rule    string   // "required", "min", "email", etc.
    Message string   // Human-readable message
    Cause   error    // Underlying error
}

type ValidationErrors struct {
    Errors []ValidationError
}
```

### Template Messages

Custom messages with interpolation:

```go
poxxy.Required().WithMessage("{field} is required")
poxxy.Min(18).WithMessage("{field} must be at least {min}, got {value}")
```

Available placeholders: `{field}`, `{value}`, `{min}`, `{max}`, `{error}`, `{expected}`

---

## Internationalization (i18n)

Built-in i18n using `golang.org/x/text`:

```go
import "golang.org/x/text/message"

// Set up translations
printer := message.NewPrinter(language.French)
schema := poxxy.NewSchema(...).WithPrinter(printer)

// Messages are translated automatically
```

Default messages are in English. User provides translations via x/text catalog.

---

## Input/Output

### Decoder Interface (Extensible)

```go
type Decoder interface {
    Decode(v any) error
}

// Built-in: JSON decoder
schema.Parse(jsonReader) // Uses json.NewDecoder internally

// Custom decoder
schema.ParseDecoder(myYAMLDecoder)
```

### Adapters for HTTP

Adapters convert HTTP requests to the internal format (separate package):

```go
import "github.com/arkan/poxxy/v2/adapters/http"

// Adapter pattern - core stays agnostic
decoder := http.FormDecoder(r) // or http.JSONDecoder(r), http.QueryDecoder(r)
schema.ParseDecoder(decoder)
```

---

## Type Conversion

### Configurable Strictness

```go
// Global strict mode
schema := poxxy.NewSchema(...).StrictTypes(true)

// Per-field strict mode
poxxy.Field("count", &count).StrictType()

// Default: flexible conversion (string -> int, float -> int, etc.)
```

Flexible mode uses internal conversion (no external go-convert dependency).

---

## Default Values

### Configurable Empty Handling

```go
// Apply default only if field is absent from input
poxxy.Field("status", &status).Default("pending")

// Apply default if field is absent OR empty
poxxy.Field("name", &name).DefaultOnEmpty("Anonymous")
```

---

## Nested Structures

### Struct Fields (Callback Builder)

```go
poxxy.Struct("address", &user.Address, func(s *poxxy.Schema) {
    s.Field("street", &user.Address.Street).Validate(poxxy.Required())
    s.Field("city", &user.Address.City).Validate(poxxy.Required())
    s.Field("zip", &user.Address.Zip).Validate(poxxy.Pattern(`^\d{5}$`))
})
```

### Slice of Objects (Callback Per Item)

```go
poxxy.Slice("users", &users, func(u *User, s *poxxy.Schema) {
    s.Field("name", &u.Name).Validate(poxxy.Required())
    s.Field("email", &u.Email).Validate(poxxy.Required(), poxxy.Email())
})
```

### Circular Reference Detection

Poxxy detects circular references at schema construction and returns an error:

```go
// Will error: "circular reference detected: User -> Address -> User"
```

---

## Pointer Fields - Null vs Absent

After `Parse()`, pointer fields expose methods to differentiate:

```go
poxxy.Pointer("nickname", &user.Nickname)

// After parsing
field := schema.GetField("nickname")
field.IsAbsent() // true if key not in JSON
field.IsNull()   // true if key present but value is null
```

---

## Performance

### Minimal Reflection

V2 minimizes reflection usage:
- Generic types eliminate most runtime type checks
- Field construction uses generics, not reflection
- Validation uses typed interfaces

### Not Thread-Safe

Schemas are **not thread-safe**. Create a new schema per request/goroutine:

```go
// Each handler creates its own schema
func handler(w http.ResponseWriter, r *http.Request) {
    var user User
    schema := poxxy.NewSchema(...)
    err := schema.Parse(r.Body)
}
```

---

## Debugging

### Verbose Trace via slog

```go
import "log/slog"

logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
schema := poxxy.NewSchema(...).WithLogger(logger)

// Traces each step:
// DEBUG poxxy: assigning field="email" value="test@example.com"
// DEBUG poxxy: transforming field="email" transformer="TrimSpace"
// DEBUG poxxy: validating field="email" validator="Required"
// DEBUG poxxy: validating field="email" validator="Email"
```

---

## Project Structure

Minimal file organization:

```
poxxy/
├── core.go           # Types: Schema, Field, FieldBuilder, interfaces
├── validators.go     # All built-in validators
├── transformers.go   # All built-in transformers
├── errors.go         # ValidationError, ValidationErrors
├── i18n.go           # Internationalization support
├── debug.go          # Logging/tracing
└── adapters/
    └── http/         # HTTP adapters (separate package)
        ├── json.go
        ├── form.go
        └── query.go
```

---

## Dependencies

### Minimal Dependencies

- **Core**: Only `golang.org/x/text` for i18n
- **No external conversion library** - internal implementation
- **adapters/http**: No additional dependencies (uses stdlib)

---

## Documentation Generation (Separate Package)

```go
import "github.com/arkan/poxxy/v2/openapi"
import "github.com/arkan/poxxy/v2/jsonschema"

// Generate OpenAPI spec
spec := openapi.FromSchema(schema)

// Generate JSON Schema
jsonSchema := jsonschema.FromSchema(schema)
```

---

## Testing

Table-driven tests with exhaustive coverage:

```go
func TestValidators(t *testing.T) {
    tests := []struct {
        name      string
        validator Validator[string]
        value     string
        wantErr   bool
    }{
        {"required empty", Required[string](), "", true},
        {"required valid", Required[string](), "hello", false},
        {"email invalid", Email(), "notanemail", true},
        {"email valid", Email(), "test@example.com", false},
        // ... many more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.validator.Validate(tt.value)
            if (err != nil) != tt.wantErr {
                t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Example Usage

```go
package main

import (
    "net/http"
    "github.com/arkan/poxxy/v2"
)

type CreateUserRequest struct {
    Name    string
    Email   string
    Age     int
    Address Address
}

type Address struct {
    Street string
    City   string
    Zip    string
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest

    schema := poxxy.NewSchema(
        poxxy.Field("name", &req.Name).
            Transform(poxxy.TrimSpace()).
            Validate(poxxy.Required(), poxxy.MinLength(2)),

        poxxy.Field("email", &req.Email).
            Transform(poxxy.TrimSpace(), poxxy.ToLower()).
            Validate(poxxy.Required(), poxxy.Email()),

        poxxy.Field("age", &req.Age).
            Validate(poxxy.Min(0), poxxy.Max(150)),

        poxxy.Struct("address", &req.Address, func(s *poxxy.Schema) {
            s.Field("street", &req.Address.Street).Validate(poxxy.Required())
            s.Field("city", &req.Address.City).Validate(poxxy.Required())
            s.Field("zip", &req.Address.Zip).Validate(poxxy.Pattern(`^\d{5}$`))
        }),
    )

    if err := schema.Parse(r.Body); err != nil {
        if ve, ok := err.(*poxxy.ValidationErrors); ok {
            // Handle validation errors
            respondWithValidationErrors(w, ve)
            return
        }
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // req is now populated and validated
    createUser(req)
}
```

---

## Summary of Decisions

| Aspect | Decision |
|--------|----------|
| API Style | Fluent (method chaining) |
| Conditional Validation | Cross-field validators with schema access |
| Template Override | No override - templates are immutable |
| Type Constraints | Interface constraints where possible, type inference from FieldBuilder |
| Doc Generation | Separate packages (poxxy/openapi, poxxy/jsonschema) |
| HTTP Support | Adapter pattern - core is source-agnostic |
| Error Paths | Flat field names ("users[2].email") |
| Union Types | Dropped - user handles manually |
| Transformers | Typed with compile-time constraints |
| Type Conversion | Configurable (strict or flexible per field/schema) |
| Empty Value Handling | Configurable (DefaultOnEmpty vs Default) |
| Immutability | Not a priority - mutable schemas |
| Field References | String names |
| Field Types Kept | Value, Pointer, Slice, Map, Struct, Convert |
| Field Types Dropped | Array, NestedMap, Union, HTTPMap |
| Error Messages | Template with interpolation |
| i18n | Built-in with golang.org/x/text |
| Performance | Minimize reflection |
| Concurrency | Not thread-safe |
| Async Validation | Callback pattern (deferred functions) |
| Dependencies | Minimal - only x/text in core |
| Testing | Table-driven |
| Migration | Breaking change (poxxy/v2) |
| Input Type | io.Reader with Decoder interface |
| Input Formats | Extensible via Decoder interface |
| Main Method | Parse() |
| Custom Validators | Helper function (ValidatorFunc) |
| Project Structure | Minimal files (core.go, validators.go, transformers.go) |
| Nested Structs | Callback builder |
| Slice Items | Callback per item |
| Circular References | Detect + error |
| Null vs Absent | Separate methods (IsAbsent, IsNull) |
| Debugging | Verbose trace via slog |
