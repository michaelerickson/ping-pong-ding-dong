package main

import (
	"log"
	"os"
	"strings"
)

// serviceStatus represents the health of the service
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
		log.Fatalf("Unknown mode: %s", mode)
	}
	log.Printf("Running in mode: %s", mode)
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
