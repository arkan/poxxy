package main

import (
	"fmt"
	"net/http"

	"github.com/arkan/poxxy"
)

func main() {
	type Attachment struct {
		Name string
		Size int64
	}

	// var attachments []Attachment
	var attachments map[int64]Attachment

	schema := poxxy.NewSchema(
		// poxxy.SliceOf[Attachment]("attachments", &attachments, nil, poxxy.WithValidators(poxxy.Required())),
		poxxy.Map[int64, Attachment]("attachments", &attachments, nil, poxxy.WithValidators(poxxy.Required())),
	)

	http.HandleFunc("POST /files", func(w http.ResponseWriter, r *http.Request) {
		if err := schema.ApplyHTTPRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Attachments: %v", attachments)))
	})

	fmt.Println("Server is running on port 8080")
	fmt.Println("Try to send a POST request to http://localhost:8080/info with the following curl request:")
	fmt.Println(`curl -X POST http://localhost:8080/files -H "Content-Type: application/x-www-form-urlencoded"  -d 'attachments[0][name]=file-abc.pdf&attachments[0][size]=1234&attachments[1][name]=file-xyz.pdf&attachments[1][size]=3456'`)

	http.ListenAndServe(":8080", nil)
}
