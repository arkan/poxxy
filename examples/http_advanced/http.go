package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arkan/poxxy"
)

type Attachment struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int    `json:"size"`
}

func main() {
	fmt.Printf("=== HTTPMap with Default Values Example ===\n")

	// Example 1: HTTPMap with default value
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
			poxxy.WithSchema(s, poxxy.Value("url", &a.URL, poxxy.WithValidators(poxxy.Required())))
			poxxy.WithSchema(s, poxxy.Value("filename", &a.Filename, poxxy.WithValidators(poxxy.Required())))
			poxxy.WithSchema(s, poxxy.Value("size", &a.Size, poxxy.WithDefault(0)))
		}, poxxy.WithDefault(defaultAttachments)),
	)

	// Test with empty form data - should use default
	fmt.Println("1. Testing with empty form data (should use default):")
	emptyData := map[string]interface{}{}
	err := schema.Apply(emptyData)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Attachments: %+v\n", attachments)
	}

	// Test with form data - should override default
	fmt.Println("\n2. Testing with form data (should override default):")
	formData := map[string]interface{}{
		"attachments[0][url]":      "https://example.com/doc1.pdf",
		"attachments[0][filename]": "doc1.pdf",
		"attachments[0][size]":     "2048",
		"attachments[1][url]":      "https://example.com/doc2.pdf",
		"attachments[1][filename]": "doc2.pdf",
		"attachments[1][size]":     "3072",
	}

	err = schema.Apply(formData)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Attachments: %+v\n", attachments)
	}

	// Example 2: HTTPMap with complex default values
	fmt.Println("\n3. Testing with complex default values:")

	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	var users map[int]User
	defaultUsers := map[int]User{
		1: {Name: "Default User", Email: "default@example.com", Age: 25},
		2: {Name: "Admin User", Email: "admin@example.com", Age: 30},
	}

	userSchema := poxxy.NewSchema(
		poxxy.HTTPMap("users", &users, func(s *poxxy.Schema, u *User) {
			poxxy.WithSchema(s, poxxy.Value("name", &u.Name, poxxy.WithValidators(poxxy.Required())))
			poxxy.WithSchema(s, poxxy.Value("email", &u.Email, poxxy.WithValidators(poxxy.Required(), poxxy.Email())))
			poxxy.WithSchema(s, poxxy.Value("age", &u.Age, poxxy.WithValidators(poxxy.Min(18), poxxy.Max(120))))
		}, poxxy.WithDefault(defaultUsers)),
	)

	// Test with empty data - should use default
	err = userSchema.Apply(map[string]interface{}{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Users: %+v\n", users)
	}

	// Example 3: Simulating HTTP request with form data
	fmt.Println("\n4. Simulating HTTP request with form data:")

	// Create a mock HTTP request
	form := url.Values{}
	form.Add("attachments[new][url]", "https://example.com/new.pdf")
	form.Add("attachments[new][filename]", "new.pdf")
	form.Add("attachments[new][size]", "4096")

	req, _ := http.NewRequest("POST", "/upload", nil)
	req.PostForm = form

	// Apply the request data
	err = schema.ApplyHTTPRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Attachments from HTTP request: %+v\n", attachments)
	}
}
