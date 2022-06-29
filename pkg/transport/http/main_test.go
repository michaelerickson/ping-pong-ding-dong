package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestValidModes checks that all valid modes are accepted
func TestValidModes(t *testing.T) {
	var tests = []struct {
		mode   string
		expect bool
	}{
		{mode: "ping", expect: true},
		{mode: "Ping", expect: true},
		{mode: "PING", expect: true},
		{mode: "pong", expect: true},
		{mode: "ding", expect: true},
		{mode: "dong", expect: true},
	}
	for _, test := range tests {
		valid := validMode(test.mode)
		if valid != test.expect {
			t.Fatalf(`Mode %s = %t, wanted %t`, test.mode, valid, test.expect)
		}
	}
}

// TestInvalidModes checks a few invalid modes including an empty string
func TestInvalidModes(t *testing.T) {
	var tests = []struct {
		mode   string
		expect bool
	}{
		{mode: "foo", expect: false},
		{mode: "", expect: false},
	}
	for _, test := range tests {
		valid := validMode(test.mode)
		if valid != test.expect {
			t.Fatalf(`Mode %s = %t, wanted %t`, test.mode, valid, test.expect)
		}
	}
}

// TestHealthCheckHandler sees if the health check works
func TestHealthCheckHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthCheckHandler(w, r)
	result := w.Result()
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Expected error to be nil - got %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d - got %d", http.StatusOK, w.Code)
	}
	contentType := result.Header[http.CanonicalHeaderKey("Content-Type")]

	if len(contentType) != 0 && !strings.EqualFold(contentType[0], "application/json") {
		t.Fatalf("Expected JSON response, got: %s", contentType[0])
	}
	var status serviceStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		t.Fatalf("Could not parse returned JSON: %s", err)
	}
	if !strings.EqualFold(status.Status, "OK") {
		t.Fatalf("Expected status OK, got %s", status.Status)
	}
}

// TestShutdownMethod ensures that only a POST is accepted
func TestShutdownMethod(t *testing.T) {
	var tests = []struct {
		method string
		expect int
	}{
		{method: http.MethodGet, expect: http.StatusMethodNotAllowed},
		{method: http.MethodHead, expect: http.StatusMethodNotAllowed},
		{method: http.MethodPost, expect: http.StatusOK},
		{method: http.MethodPut, expect: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expect: http.StatusMethodNotAllowed},
		{method: http.MethodPatch, expect: http.StatusMethodNotAllowed},
		{method: http.MethodOptions, expect: http.StatusMethodNotAllowed},
		{method: http.MethodTrace, expect: http.StatusMethodNotAllowed},
	}
	_, cancel := context.WithCancel(context.TODO())
	shutdownFunc := shutdownHandler(cancel)
	shutdown := http.HandlerFunc(shutdownFunc)
	for _, test := range tests {
		r := httptest.NewRequest(test.method, "/shutdown", nil)
		w := httptest.NewRecorder()
		shutdown(w, r)
		if w.Code != test.expect {
			t.Fatalf("Method %s expected %d got %d", test.method, test.expect, w.Code)
		}
	}
}

// TestRootMethod ensures that only GET and POST are accepted
// Note, we don't test GET and POST explicitly because GET requires a context
// and POST triggers a call to another service
func TestRootMethod(t *testing.T) {
	var tests = []struct {
		method string
		expect int
	}{
		// {method: http.MethodGet, expect: http.StatusOK},
		{method: http.MethodHead, expect: http.StatusMethodNotAllowed},
		// {method: http.MethodPost, expect: http.StatusOK},
		{method: http.MethodPut, expect: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expect: http.StatusMethodNotAllowed},
		{method: http.MethodPatch, expect: http.StatusMethodNotAllowed},
		{method: http.MethodOptions, expect: http.StatusMethodNotAllowed},
		{method: http.MethodTrace, expect: http.StatusMethodNotAllowed},
	}
	for _, test := range tests {
		r := httptest.NewRequest(test.method, "/", nil)
		w := httptest.NewRecorder()
		rootHandler(w, r)
		if w.Code != test.expect {
			t.Fatalf("Method %s expected %d got %d", test.method, test.expect, w.Code)
		}
	}
}
