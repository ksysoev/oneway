package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	listen := os.Getenv("HTTP_LISTEN")
	if listen == "" {
		slog.Error("HTTP_LISTEN not provided")
		return
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "HTTP server is running")
	})

	// Start the server
	fmt.Printf("Starting server on port %s\n", listen)
	if err := http.ListenAndServe(listen, nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
