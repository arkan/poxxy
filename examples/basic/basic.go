package main

import (
	"fmt"

	"github.com/arkan/poxxy"
)

type Address struct {
	Street string
	City   string
}

type UserProfile struct {
	Name     string
	Email    *string  // Optional field
	Address  *Address // Optional nested struct
	Settings map[string]interface{}
}

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

func main() {
	data := map[string]interface{}{
		"profile": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com", // Optional
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "Boston",
			},
			"settings": map[string]interface{}{
				"theme": map[string]interface{}{
					"dark_mode": true,
					"font_size": 14,
				},
				"notifications": map[string]interface{}{
					"email": true,
					"push":  false,
				},
			},
		},
		"tags":          [3]string{"urgent", "work", "important"}, // Fixed-size array
		"recent_scores": []int{95, 88, 92},                        // Dynamic slice
		"document": map[string]interface{}{
			"type":       "text",
			"content":    "Hello World",
			"word_count": 2,
		},
	}

	var profile UserProfile
	var tags [3]string     // Fixed array
	var recentScores []int // Dynamic slice
	var document Document  // Union/polymorphic type

	schema := poxxy.NewSchema(
		// 1. Pointer to struct with optional fields
		poxxy.Struct[UserProfile]("profile", &profile, poxxy.WithSubSchema(func(s *poxxy.Schema, p *UserProfile) {
			poxxy.WithSchema(s, poxxy.Value[string]("name", &p.Name, poxxy.WithValidators(poxxy.Required())))

			// 2. Pointer to string (optional field)
			poxxy.WithSchema(s, poxxy.Pointer[string]("email", &p.Email, poxxy.WithValidators(poxxy.Email())))

			// 3. Pointer to struct (optional nested struct)
			poxxy.WithSchema(s, poxxy.Pointer[Address]("address", &p.Address, poxxy.WithSubSchema(func(ss *poxxy.Schema, addr *Address) {
				poxxy.WithSchema(ss, poxxy.Value[string]("street", &addr.Street, poxxy.WithValidators(poxxy.Required())))
				poxxy.WithSchema(ss, poxxy.Value[string]("city", &addr.City, poxxy.WithValidators(poxxy.Required())))
			})))
		})),

		// 4. Fixed-size array (vs slice)
		poxxy.Array[string]("tags", &tags, poxxy.WithValidators(
			poxxy.Required(),
			poxxy.Unique(),
			poxxy.Each(poxxy.MinLength(1)),
		)),

		// 5. Dynamic slice for comparison
		poxxy.Slice[int]("recent_scores", &recentScores, poxxy.WithValidators(
			poxxy.MaxLength(10),
			poxxy.Each(poxxy.Min(0), poxxy.Max(100)),
		)),

		// 6. Union/Polymorphic types
		poxxy.Union("document", &document, func(data map[string]interface{}) (interface{}, error) {
			docType, ok := data["type"].(string)
			if !ok {
				return nil, fmt.Errorf("missing or invalid document type")
			}

			switch docType {
			case "text":
				var doc TextDocument
				// Use sub-schema to validate and assign
				subSchema := poxxy.NewSchema(
					poxxy.Value[string]("content", &doc.Content, poxxy.WithValidators(poxxy.Required())),
					poxxy.Value[int]("word_count", &doc.WordCount, poxxy.WithValidators(poxxy.Min(0))),
				)
				if err := subSchema.Apply(data); err != nil {
					return nil, err
				}
				return doc, nil
			case "image":
				var doc ImageDocument
				subSchema := poxxy.NewSchema(
					poxxy.Value[string]("url", &doc.URL, poxxy.WithValidators(poxxy.Required(), poxxy.URL())),
					poxxy.Value[int]("width", &doc.Width, poxxy.WithValidators(poxxy.Min(1))),
					poxxy.Value[int]("height", &doc.Height, poxxy.WithValidators(poxxy.Min(1))),
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

	if err := schema.Apply(data); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("Profile: %+v\n", profile)
	fmt.Printf("Tags (array): %v\n", tags)
	fmt.Printf("Recent scores (slice): %v\n", recentScores)
	fmt.Printf("Document: %+v (Type: %s)\n", document, document.GetType())
}
