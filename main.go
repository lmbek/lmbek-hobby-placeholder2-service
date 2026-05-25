package main

import (
	"log/slog"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authentication Service is Healthy"))
	})

	slog.Info("Authentication Service starting", "port", 8081)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		slog.Error("Error starting service", "error", err)
	}
}
