package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Authentication Service is Healthy")
	})

	fmt.Println("Authentication Service starting on :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Printf("Error starting service: %v\n", err)
	}
}
