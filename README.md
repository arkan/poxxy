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

## Advanced Field Types

### HTTPMap Fields - HTTP Form Data Management

`HTTPMap` fields are specifically designed to handle complex HTTP form data where each map value is a structure with its own validation schema. This functionality is particularly useful for web forms with object collections.

#### Working with Form Data

`HTTPMap` fields automatically parse form data in the format `field[key][subfield]=value` and convert them to typed structures.

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

schema := poxxy.NewSchema(
    poxxy.HTTPMap("attachments", &attachments, func(s *poxxy.Schema, a *Attachment) {
        poxxy.WithSchema(s, poxxy.Value("url", &a.URL,
            poxxy.WithValidators(poxxy.Required(), poxxy.URL()),
            poxxy.WithDescription("Attachment URL"),
        ))
        poxxy.WithSchema(s, poxxy.Value("filename", &a.Filename,
            poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(1)),
            poxxy.WithTransformers(poxxy.TrimSpace()),
        ))
        poxxy.WithSchema(s, poxxy.Value("size", &a.Size,
            poxxy.WithDefault(0),
            poxxy.WithValidators(poxxy.Min(0)),
        ))
    }, poxxy.WithDefault(defaultAttachments)),
)
```

#### Form Data Handling

This schema can handle form data like:
```
attachments[0][url]=https://example.com/doc1.pdf&attachments[0][filename]=doc1.pdf&attachments[0][size]=2048
attachments[1][url]=https://example.com/doc2.pdf&attachments[1][filename]=doc2.pdf&attachments[1][size]=3072
```

#### Complete Example with HTTP Validation

```go
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
    var attachments map[string]Attachment

    schema := poxxy.NewSchema(
        poxxy.HTTPMap("attachments", &attachments, func(s *poxxy.Schema, a *Attachment) {
            poxxy.WithSchema(s, poxxy.Value("url", &a.URL,
                poxxy.WithValidators(poxxy.Required(), poxxy.URL()),
            ))
            poxxy.WithSchema(s, poxxy.Value("filename", &a.Filename,
                poxxy.WithValidators(poxxy.Required()),
                poxxy.WithTransformers(poxxy.TrimSpace()),
            ))
            poxxy.WithSchema(s, poxxy.Value("size", &a.Size,
                poxxy.WithValidators(poxxy.Min(1), poxxy.Max(10*1024*1024)), // Max 10MB
            ))
        }),
    )

    if err := schema.ApplyHTTPRequest(r); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Process validated attachments
    for key, attachment := range attachments {
        fmt.Printf("Attachment %s: %s (%d bytes)\n", key, attachment.Filename, attachment.Size)
    }
}
```

#### Advanced Use Cases

**Multiple User Management:**
```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

var users map[int]User
schema := poxxy.NewSchema(
    poxxy.HTTPMap("users", &users, func(s *poxxy.Schema, u *User) {
        poxxy.WithSchema(s, poxxy.Value("name", &u.Name,
            poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(2)),
            poxxy.WithTransformers(poxxy.TrimSpace(), poxxy.Capitalize()),
        ))
        poxxy.WithSchema(s, poxxy.Value("email", &u.Email,
            poxxy.WithValidators(poxxy.Required(), poxxy.Email()),
            poxxy.WithTransformers(poxxy.SanitizeEmail()),
        ))
        poxxy.WithSchema(s, poxxy.Value("age", &u.Age,
            poxxy.WithValidators(poxxy.Min(18), poxxy.Max(120)),
        ))
    }),
)
```

### NestedMap Fields - Nested Map Validation

`NestedMap` fields allow you to validate each key-value pair of a map with custom rules. Unlike `HTTPMap` fields that handle structures, `NestedMap` fields work with simple values.

#### Main Use Cases

**Configuration Parameters Validation:**
```go
var config map[string]string
schema := poxxy.NewSchema(
    poxxy.NestedMap("config", &config, func(s *poxxy.Schema, key string, value *string) {
        // Key validation
        poxxy.WithSchema(s, poxxy.ValueWithoutAssign("key",
            poxxy.WithValidators(
                poxxy.Required(),
                poxxy.In("theme", "language", "timezone", "currency"),
            ),
        ))
        // Value validation
        poxxy.WithSchema(s, poxxy.ValueWithoutAssign("value",
            poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(1)),
        ))
    }),
)
```

**Metadata Validation:**
```go
var metadata map[string]interface{}
schema := poxxy.NewSchema(
    poxxy.NestedMap("metadata", &metadata, func(s *poxxy.Schema, key string, value *interface{}) {
        // Key validation (must be lowercase)
        poxxy.WithSchema(s, poxxy.ValueWithoutAssign("key",
            poxxy.WithValidators(
                poxxy.Required(),
                poxxy.ValidatorFunc(func(key string, fieldName string) error {
                    if key != strings.ToLower(key) {
                        return fmt.Errorf("key must be lowercase")
                    }
                    return nil
                }),
            ),
        ))
        // Value validation (must be string or number)
        poxxy.WithSchema(s, poxxy.ValueWithoutAssign("value",
            poxxy.WithValidators(
                poxxy.Required(),
                poxxy.ValidatorFunc(func(value interface{}, fieldName string) error {
                    switch v := value.(type) {
                    case string, int, float64:
                        return nil
                    default:
                        return fmt.Errorf("value must be string, int, or float, got %T", v)
                    }
                }),
            ),
        ))
    }),
)
```

**Permissions Validation:**
```go
var permissions map[string]bool
schema := poxxy.NewSchema(
    poxxy.NestedMap("permissions", &permissions, func(s *poxxy.Schema, key string, value *bool) {
        // Key validation (must be a valid permission name)
        poxxy.WithSchema(s, poxxy.ValueWithoutAssign("key",
            poxxy.WithValidators(
                poxxy.Required(),
                poxxy.ValidatorFunc(func(key string, fieldName string) error {
                    validPermissions := []string{"read", "write", "delete", "admin"}
                    for _, perm := range validPermissions {
                        if key == perm {
                            return nil
                        }
                    }
                    return fmt.Errorf("invalid permission: %s", key)
                }),
            ),
        ))
    }),
)
```

### Union Fields - Polymorphism Patterns

`Union` fields allow you to handle polymorphic structures where the exact type is determined by a discriminator. This functionality is ideal for APIs that need to handle different types of objects.

#### Polymorphism Patterns

**Polymorphic Documents:**
```go
type Document interface {
    GetType() string
}

type TextDocument struct {
    Content   string `json:"content"`
    WordCount int    `json:"word_count"`
    Language  string `json:"language"`
}

type ImageDocument struct {
    URL      string `json:"url"`
    Width    int    `json:"width"`
    Height   int    `json:"height"`
    Format   string `json:"format"`
}

type VideoDocument struct {
    URL       string `json:"url"`
    Duration  int    `json:"duration"`
    Quality   string `json:"quality"`
    Subtitles bool   `json:"subtitles"`
}

// GetType method implementations
func (t TextDocument) GetType() string { return "text" }
func (i ImageDocument) GetType() string { return "image" }
func (v VideoDocument) GetType() string { return "video" }

var document Document
schema := poxxy.NewSchema(
    poxxy.Union("document", &document, func(data map[string]interface{}) (interface{}, error) {
        docType, ok := data["type"].(string)
        if !ok {
            return nil, fmt.Errorf("missing or invalid document type")
        }

        switch docType {
        case "text":
            var doc TextDocument
            subSchema := poxxy.NewSchema(
                poxxy.Value("content", &doc.Content,
                    poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(10)),
                    poxxy.WithTransformers(poxxy.TrimSpace()),
                ),
                poxxy.Value("word_count", &doc.WordCount,
                    poxxy.WithValidators(poxxy.Min(0)),
                ),
                poxxy.Value("language", &doc.Language,
                    poxxy.WithDefault("en"),
                    poxxy.WithValidators(poxxy.In("en", "fr", "es", "de")),
                ),
            )
            if err := subSchema.Apply(data); err != nil {
                return nil, err
            }
            return doc, nil

        case "image":
            var doc ImageDocument
            subSchema := poxxy.NewSchema(
                poxxy.Value("url", &doc.URL,
                    poxxy.WithValidators(poxxy.Required(), poxxy.URL()),
                ),
                poxxy.Value("width", &doc.Width,
                    poxxy.WithValidators(poxxy.Min(1), poxxy.Max(10000)),
                ),
                poxxy.Value("height", &doc.Height,
                    poxxy.WithValidators(poxxy.Min(1), poxxy.Max(10000)),
                ),
                poxxy.Value("format", &doc.Format,
                    poxxy.WithDefault("jpeg"),
                    poxxy.WithValidators(poxxy.In("jpeg", "png", "gif", "webp")),
                ),
            )
            if err := subSchema.Apply(data); err != nil {
                return nil, err
            }
            return doc, nil

        case "video":
            var doc VideoDocument
            subSchema := poxxy.NewSchema(
                poxxy.Value("url", &doc.URL,
                    poxxy.WithValidators(poxxy.Required(), poxxy.URL()),
                ),
                poxxy.Value("duration", &doc.Duration,
                    poxxy.WithValidators(poxxy.Min(1)),
                ),
                poxxy.Value("quality", &doc.Quality,
                    poxxy.WithDefault("720p"),
                    poxxy.WithValidators(poxxy.In("480p", "720p", "1080p", "4k")),
                ),
                poxxy.Value("subtitles", &doc.Subtitles,
                    poxxy.WithDefault(false),
                ),
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

**Polymorphic Notifications:**
```go
type Notification interface {
    GetChannel() string
}

type EmailNotification struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

type SMSNotification struct {
    Phone   string `json:"phone"`
    Message string `json:"message"`
}

type PushNotification struct {
    Token   string            `json:"token"`
    Title   string            `json:"title"`
    Body    string            `json:"body"`
    Data    map[string]string `json:"data"`
}

func (e EmailNotification) GetChannel() string { return "email" }
func (s SMSNotification) GetChannel() string { return "sms" }
func (p PushNotification) GetChannel() string { return "push" }

var notification Notification
schema := poxxy.NewSchema(
    poxxy.Union("notification", &notification, func(data map[string]interface{}) (interface{}, error) {
        channel, ok := data["channel"].(string)
        if !ok {
            return nil, fmt.Errorf("missing or invalid notification channel")
        }

        switch channel {
        case "email":
            var notif EmailNotification
            subSchema := poxxy.NewSchema(
                poxxy.Value("to", &notif.To,
                    poxxy.WithValidators(poxxy.Required(), poxxy.Email()),
                    poxxy.WithTransformers(poxxy.SanitizeEmail()),
                ),
                poxxy.Value("subject", &notif.Subject,
                    poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(1), poxxy.MaxLength(200)),
                ),
                poxxy.Value("body", &notif.Body,
                    poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(10)),
                ),
            )
            if err := subSchema.Apply(data); err != nil {
                return nil, err
            }
            return notif, nil

        case "sms":
            var notif SMSNotification
            subSchema := poxxy.NewSchema(
                poxxy.Value("phone", &notif.Phone,
                    poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(10)),
                ),
                poxxy.Value("message", &notif.Message,
                    poxxy.WithValidators(poxxy.Required(), poxxy.MaxLength(160)),
                ),
            )
            if err := subSchema.Apply(data); err != nil {
                return nil, err
            }
            return notif, nil

        case "push":
            var notif PushNotification
            subSchema := poxxy.NewSchema(
                poxxy.Value("token", &notif.Token,
                    poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(32)),
                ),
                poxxy.Value("title", &notif.Title,
                    poxxy.WithValidators(poxxy.Required(), poxxy.MaxLength(50)),
                ),
                poxxy.Value("body", &notif.Body,
                    poxxy.WithValidators(poxxy.Required(), poxxy.MaxLength(200)),
                ),
                poxxy.Map("data", &notif.Data),
            )
            if err := subSchema.Apply(data); err != nil {
                return nil, err
            }
            return notif, nil

        default:
            return nil, fmt.Errorf("unknown notification channel: %s", channel)
        }
    }),
)
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

Custom transformers allow you to create data transformations specific to your business domain.

**Basic Transformers:**
```go
// Phone number normalization transformer
phoneTransformer := poxxy.CustomTransformer(func(phone string) (string, error) {
    // Remove all non-numeric characters
    cleaned := strings.Map(func(r rune) rune {
        if r >= '0' && r <= '9' {
            return r
        }
        return -1
    }, phone)

    // Check length
    if len(cleaned) < 10 {
        return "", fmt.Errorf("phone number too short")
    }

    // Format as international format
    if len(cleaned) == 10 {
        return "+1" + cleaned, nil
    }

    return "+" + cleaned, nil
})

var phone string
schema := poxxy.NewSchema(
    poxxy.Value("phone", &phone,
        poxxy.WithTransformers(phoneTransformer),
        poxxy.WithValidators(poxxy.Required()),
    ),
)
```

**Business Validation Transformers:**
```go
// Age validation transformer with automatic calculation
ageTransformer := poxxy.CustomTransformer(func(birthDate string) (int, error) {
    birth, err := time.Parse("2006-01-02", birthDate)
    if err != nil {
        return 0, fmt.Errorf("invalid birth date format")
    }

    age := time.Now().Year() - birth.Year()
    if time.Now().YearDay() < birth.YearDay() {
        age--
    }

    if age < 0 {
        return 0, fmt.Errorf("birth date cannot be in the future")
    }

    return age, nil
})

var age int
schema := poxxy.NewSchema(
    poxxy.Convert("birth_date", &age, func(birthDate string) (int, error) {
        return ageTransformer.Transform(birthDate)
    }, poxxy.WithValidators(poxxy.Min(18))),
)
```

**Formatting Transformers:**
```go
// Currency formatting transformer
currencyTransformer := poxxy.CustomTransformer(func(amount float64) (string, error) {
    if amount < 0 {
        return "", fmt.Errorf("amount cannot be negative")
    }
    return fmt.Sprintf("$%.2f", amount), nil
})

var formattedAmount string
schema := poxxy.NewSchema(
    poxxy.Convert("amount", &formattedAmount, func(amount float64) (string, error) {
        return currencyTransformer.Transform(amount)
    }),
)
```

**Security Transformers:**
```go
// HTML sanitization transformer
htmlSanitizer := poxxy.CustomTransformer(func(content string) (string, error) {
    // Remove dangerous HTML tags
    dangerousTags := []string{"<script>", "</script>", "<iframe>", "</iframe>", "<object>", "</object>"}
    cleaned := content
    for _, tag := range dangerousTags {
        cleaned = strings.ReplaceAll(cleaned, tag, "")
    }

    // Limit length
    if len(cleaned) > 10000 {
        return "", fmt.Errorf("content too long")
    }

    return cleaned, nil
})

var safeContent string
schema := poxxy.NewSchema(
    poxxy.Value("content", &safeContent,
        poxxy.WithTransformers(htmlSanitizer),
        poxxy.WithValidators(poxxy.Required(), poxxy.MinLength(10)),
    ),
)
```

**Normalization Transformers:**
```go
// URL normalization transformer
urlNormalizer := poxxy.CustomTransformer(func(url string) (string, error) {
    url = strings.TrimSpace(url)

    // Add protocol if missing
    if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
        url = "https://" + url
    }

    // Normalize to lowercase
    url = strings.ToLower(url)

    return url, nil
})

var normalizedURL string
schema := poxxy.NewSchema(
    poxxy.Value("url", &normalizedURL,
        poxxy.WithTransformers(urlNormalizer),
        poxxy.WithValidators(poxxy.Required(), poxxy.URL()),
    ),
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

### Complex Validations with Each(), Unique(), etc.

Complex validations allow you to validate collections with sophisticated rules.

#### Each() - Validating Each Element

**Email List Validation:**
```go
var emails []string
schema := poxxy.NewSchema(
    poxxy.Slice("emails", &emails,
        poxxy.WithValidators(
            poxxy.Required(),
            poxxy.Each(poxxy.Email(), poxxy.MinLength(5)),
            poxxy.Unique(),
        ),
        poxxy.WithTransformers(
            poxxy.CustomTransformer(func(emails []string) ([]string, error) {
                // Normalize all emails
                for i, email := range emails {
                    emails[i] = strings.ToLower(strings.TrimSpace(email))
                }
                return emails, nil
            }),
        ),
    ),
)
```

**User List Validation:**
```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

var users []User
schema := poxxy.NewSchema(
    poxxy.Slice("users", &users,
        poxxy.WithValidators(
            poxxy.Required(),
            poxxy.Each(
                poxxy.ValidatorFunc(func(user User, fieldName string) error {
                    if user.Name == "" {
                        return fmt.Errorf("name is required")
                    }
                    if user.Email == "" {
                        return fmt.Errorf("email is required")
                    }
                    if user.Age < 18 {
                        return fmt.Errorf("user must be at least 18 years old")
                    }
                    return nil
                }),
            ),
            poxxy.UniqueBy(func(user interface{}) interface{} {
                return user.(User).Email
            }),
        ),
    ),
)
```

**Score Array Validation:**
```go
var scores [5]int
schema := poxxy.NewSchema(
    poxxy.Array("scores", &scores,
        poxxy.WithValidators(
            poxxy.Each(poxxy.Min(0), poxxy.Max(100)),
            poxxy.ValidatorFunc(func(scores [5]int, fieldName string) error {
                // Check that at least 3 scores are above 50
                count := 0
                for _, score := range scores {
                    if score > 50 {
                        count++
                    }
                }
                if count < 3 {
                    return fmt.Errorf("at least 3 scores must be above 50")
                }
                return nil
            }),
        ),
    ),
)
```

#### Unique() and UniqueBy() - Uniqueness Validation

**Simple Uniqueness Validation:**
```go
var tags []string
schema := poxxy.NewSchema(
    poxxy.Slice("tags", &tags,
        poxxy.WithValidators(
            poxxy.Required(),
            poxxy.Each(poxxy.MinLength(2), poxxy.MaxLength(20)),
            poxxy.Unique(),
        ),
        poxxy.WithTransformers(
            poxxy.CustomTransformer(func(tags []string) ([]string, error) {
                // Normalize tags (lowercase, no spaces)
                for i, tag := range tags {
                    tags[i] = strings.ToLower(strings.ReplaceAll(tag, " ", "-"))
                }
                return tags, nil
            }),
        ),
    ),
)
```

**Uniqueness Validation by Property:**
```go
type Product struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Price float64 `json:"price"`
}

var products []Product
schema := poxxy.NewSchema(
    poxxy.Slice("products", &products,
        poxxy.WithValidators(
            poxxy.Required(),
            poxxy.Each(
                poxxy.ValidatorFunc(func(product Product, fieldName string) error {
                    if product.ID == "" {
                        return fmt.Errorf("product ID is required")
                    }
                    if product.Name == "" {
                        return fmt.Errorf("product name is required")
                    }
                    if product.Price <= 0 {
                        return fmt.Errorf("product price must be positive")
                    }
                    return nil
                }),
            ),
            // Check uniqueness by ID
            poxxy.UniqueBy(func(product interface{}) interface{} {
                return product.(Product).ID
            }),
        ),
    ),
)
```

**Uniqueness Validation in a Map:**
```go
var settings map[string]string
schema := poxxy.NewSchema(
    poxxy.Map("settings", &settings,
        poxxy.WithValidators(
            poxxy.Required(),
            poxxy.Unique(), // Check uniqueness of values
            poxxy.WithMapKeys("theme", "language", "timezone"), // Required keys
        ),
    ),
)
```

#### Complex Cross-Field Validations

**Field Consistency Validation:**
```go
type Order struct {
    Items      []OrderItem `json:"items"`
    Total      float64     `json:"total"`
    Currency   string      `json:"currency"`
    Discount   float64     `json:"discount"`
    FinalTotal float64     `json:"final_total"`
}

type OrderItem struct {
    ProductID string  `json:"product_id"`
    Quantity  int     `json:"quantity"`
    Price     float64 `json:"price"`
    Subtotal  float64 `json:"subtotal"`
}

var order Order
schema := poxxy.NewSchema(
    poxxy.Struct("order", &order, poxxy.WithSubSchema(func(s *poxxy.Schema, o *Order) {
        poxxy.WithSchema(s, poxxy.Slice("items", &o.Items,
            poxxy.WithValidators(
                poxxy.Required(),
                poxxy.Each(
                    poxxy.ValidatorFunc(func(item OrderItem, fieldName string) error {
                        if item.ProductID == "" {
                            return fmt.Errorf("product ID is required")
                        }
                        if item.Quantity <= 0 {
                            return fmt.Errorf("quantity must be positive")
                        }
                        if item.Price <= 0 {
                            return fmt.Errorf("price must be positive")
                        }
                        // Check that the subtotal is correct
                        expectedSubtotal := float64(item.Quantity) * item.Price
                        if item.Subtotal != expectedSubtotal {
                            return fmt.Errorf("subtotal mismatch: expected %.2f, got %.2f",
                                expectedSubtotal, item.Subtotal)
                        }
                        return nil
                    }),
                ),
                poxxy.UniqueBy(func(item interface{}) interface{} {
                    return item.(OrderItem).ProductID
                }),
            ),
        ))

        poxxy.WithSchema(s, poxxy.Value("total", &o.Total,
            poxxy.WithValidators(poxxy.Min(0)),
        ))

        poxxy.WithSchema(s, poxxy.Value("currency", &o.Currency,
            poxxy.WithValidators(poxxy.In("USD", "EUR", "GBP")),
        ))

        poxxy.WithSchema(s, poxxy.Value("discount", &o.Discount,
            poxxy.WithValidators(poxxy.Min(0), poxxy.Max(100)),
        ))

        poxxy.WithSchema(s, poxxy.Value("final_total", &o.FinalTotal,
            poxxy.WithValidators(
                poxxy.ValidatorFunc(func(finalTotal float64, fieldName string) error {
                    // Check that the final total is consistent
                    expectedTotal := o.Total * (1 - o.Discount/100)
                    if finalTotal != expectedTotal {
                        return fmt.Errorf("final total mismatch: expected %.2f, got %.2f",
                            expectedTotal, finalTotal)
                    }
                    return nil
                }),
            ),
        ))
    })),
)
```

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


