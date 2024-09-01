package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	Timeout = 5 * time.Second
)

func main() {
	listen := os.Getenv("HTTP_LISTEN")
	if listen == "" {
		slog.Error("HTTP_LISTEN not provided")
		return
	}

	mux := http.NewServeMux()
	httpSever := &http.Server{
		Addr:         listen,
		ReadTimeout:  Timeout,
		WriteTimeout: Timeout,
		Handler:      mux,
	}

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "HTTP server is running")
	})

	// Start the server
	fmt.Printf("Starting server on port %s\n", listen)

	if err := httpSever.ListenAndServe(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
