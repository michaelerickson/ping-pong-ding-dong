package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// serviceMsg defines the types of messages we pass around
type serviceMsg struct {
	Msg string
}

// mode indicates the mode we are running in
var mode string

// main starts the service and listens for requests
func main() {
	// Log in UTC time with microsecond resolution
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds)

	// Figure out what mode we should be running in
	if m := os.Getenv("PPDD_MODE"); m == "" {
		mode = "mode not set"
	} else {
		mode = m
	}
	if !validMode(mode) {
		log.Fatalf("Unknown mode: %s, check PPDD_MODE", mode)
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	log.Printf("Running in mode: %s on port %s", mode, httpPort)

	m := http.NewServeMux()
	s := http.Server{Addr: ":" + httpPort, Handler: m}

	// Establish a context so we can shut things down cleanly
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle the /shutdown endpoint which should gracefully shut down the
	// service by canceling the context.
	shutdownFunc := shutdownHandler(cancel)

	// Add handlers to the mux
	shutdown := http.HandlerFunc(shutdownFunc)
	health := http.HandlerFunc(healthCheckHandler)
	m.Handle("/shutdown", loggingMiddleware(shutdown))
	m.Handle("/health", loggingMiddleware(health))

	// Launch the server in a go routing. This way the main thread can listen
	// for the context being canceled and gracefully shut things down.
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	// Listen for the canceled context and shut things down cleanly
	select {
	case <-ctx.Done():
		if err := s.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down server: %s", err)
		}
	}
	log.Println("Application finished")
}

// validMode determines if the mode is one that we support
func validMode(m string) bool {
	validModes := getModes()
	for _, s := range validModes {
		// Using EqualFold gives is a case-insensitive compare
		if strings.EqualFold(m, s) {
			return true
		}
	}
	return false
}

// getModes returns the valid values of PPDD_MODE
func getModes() []string {
	modes := []string{"ping", "pong", "ding", "dong"}
	return modes
}

// serviceStatus represents the health of our little service
type serviceStatus struct {
	Status string
}

// healthCheckHandler handles requests to `/health`
func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	status := serviceStatus{Status: "OK"}
	response, err := json.Marshal(status)
	if err != nil {
		log.Printf("JSON error: %s", err)
		http.Error(w, "JSON error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

// shutdownHandler handles posts to `/shutdown`
func shutdownHandler(cancel context.CancelFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.EqualFold(r.Method, http.MethodPost) {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if _, err := w.Write([]byte("Shutting down\n")); err != nil {
			log.Printf("Error writing response: %s", err)
		}
		cancel()
	}
}

// loggingMiddleware logs all requests to the service
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
