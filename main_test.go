package main

import "testing"

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
