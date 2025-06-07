package main

import (
	"fmt"
	"net/http"

	"github.com/arkan/poxxy"
)

func main() {
	var name string
	var age int64
	var isAdmin bool

	schema := poxxy.NewSchema(
		poxxy.Value[string]("name", &name, poxxy.WithValidators(poxxy.Required())),
		poxxy.Value[int64]("age", &age, poxxy.WithValidators(poxxy.Required(), poxxy.Min(18), poxxy.Max(100))),
		poxxy.Value[bool]("is_admin", &isAdmin, poxxy.WithValidators(poxxy.Required())),
	)

	http.HandleFunc("POST /info", func(w http.ResponseWriter, r *http.Request) {
		if err := schema.ApplyHTTPRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hello, %s! You are %d years old.", name, age)))
	})

	fmt.Println("Server is running on port 8080")
	fmt.Println("Try to send a POST request to http://localhost:8080/info with the following curl request:")
	fmt.Println(`curl -X POST http://localhost:8080/info -H "Content-Type: application/x-www-form-urlencoded"  -d 'name=John&age=25&is_admin=true'`)

	http.ListenAndServe(":8080", nil)
}
