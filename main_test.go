package main_test

import (
	"errors"
	"testing"
)

func TestURLShortner(t *testing.T) {

	var tests = []struct {
		name     string
		in       string
		expected error
	}{
		{"NoScheme", "google.com", nil},
		{"WithScheme", "https://google.com", nil},
		{"NoSchemeAndTLD", "google", errors.New("invalid")},
		{"NoTLD", "https://google", errors.New("invalid")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s URLShortner{}
		})
	}
}
