package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLShortner(t *testing.T) {

	var tests = []struct {
		name     string
		in       string
		expected string
		err      error
	}{
		{"NoScheme", "google.com", "1", nil},
		{"WithScheme", "https://google.com", "1", nil},
		{"NoSchemeAndTLD", "google", "", errors.New("invalid")},
		{"NoTLD", "https://google", "", errors.New("invalid")},
	}

	for _, test := range tests {
		t.Run("Shorten"+test.name, func(t *testing.T) {
			s := &URLShortner{i: 0, store: make(map[string]string)}

			resp, err := s.Shorten(test.in)

			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expected, resp)
		})
	}

	var tests2 = []struct {
		name     string
		in1      string
		in2      string
		expected string
		err      error
	}{
		{"NoScheme", "google.com", "1", "https://google.com", nil},
		{"WithScheme", "https://google.com", "1", "https://google.com", nil},
		{"NoSchemeAndTLD", "google", "1", "", errors.New("Not found")},
		{"NoTLD", "https://google", "1", "", errors.New("Not found")},
	}

	for _, test := range tests2 {
		t.Run("Get"+test.name, func(t *testing.T) {
			s := &URLShortner{i: 0, store: make(map[string]string)}

			s.Shorten(test.in1)
			resp, err := s.Get(test.in2)

			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expected, resp)
		})
	}
}
