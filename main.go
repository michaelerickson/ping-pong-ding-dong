package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
)

// serviceStatus represents the health of our little service
type serviceStatus struct {
	Status string
}

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
	root := http.HandlerFunc(rootHandler)
	health := http.HandlerFunc(healthCheckHandler)
	shutdown := http.HandlerFunc(shutdownFunc)
	m.Handle("/", loggingMiddleware(root))
	m.Handle("/health", loggingMiddleware(health))
	m.Handle("/shutdown", loggingMiddleware(shutdown))

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

// rootHandler deals with requests to `/`
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if strings.EqualFold(r.Method, http.MethodGet) {
		getRoot(w, r)
	} else if strings.EqualFold(r.Method, http.MethodPost) {
		postRoot(w, r)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// getRoot handles GET requests to `/`
func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	srvAddr := ctx.Value(http.LocalAddrContextKey).(net.Addr)
	response := fmt.Sprintf("Hello from: %s:\n", srvAddr)
	response += fmt.Sprintf("  mode: %s\n", mode)
	response += fmt.Sprintf("  operating system: %s\n", runtime.GOOS)
	response += fmt.Sprintf("  architecture: %s\n", runtime.GOARCH)
	response += fmt.Sprintf("  number of CPUs: %d\n", runtime.NumCPU())
	if hostname, err := os.Hostname(); err != nil {
		response += "  hostname: unknown\n"
	} else {
		response += fmt.Sprintf("  hostname: %s\n", hostname)
	}
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	mBytes := float64(memStats.Sys) / (1 << 20)
	response += fmt.Sprintf("  memory MiB: %f\n", mBytes)
	env := os.Environ()
	sort.Strings(env)
	response += "Environment:\n"
	for _, e := range env {
		response += fmt.Sprintf("  %s\n", e)
	}
	response += "\n\n"
	_, err := w.Write([]byte(response))
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

// postRoot handles POST requests to `/`
func postRoot(w http.ResponseWriter, r *http.Request) {
	// Make sure we are dealing with JSON
	contentType := r.Header[http.CanonicalHeaderKey("Content-Type")]
	if len(contentType) == 0 || !strings.EqualFold(contentType[0], "application/json") {
		log.Println("Error: request not JSON")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	var msg serviceMsg
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("Error: cannot decode JSON: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if msg.Msg == "" {
		log.Println("Error: request missing `msg` key")
		http.Error(w, "No message", http.StatusBadRequest)
		return
	}
	log.Printf("Received %+v", msg)
	switch mode {
	case "ping":
		log.Println("Acting as ping")
	case "pong":
		log.Println("Acting as pong")
	case "ding":
		log.Println("Acting as ding")
	case "dong":
		log.Println("Acting as dong")
	default:
		log.Println("Mode is not set properly, doing nothing...")
	}
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
