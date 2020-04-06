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
		{"NoScheme", "google.com", "urls.com/1", nil},
		{"WithScheme", "https://google.com", "urls.com/1", nil},
		{"NoSchemeAndTLD", "google", "", errors.New("invalid")},
		{"NoTLD", "https://google", "", errors.New("invalid")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &URLShortner{baseURL: "urls.com", i: 0, store: make(map[string]string)}

			resp, err := s.Shorten(test.in)

			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expected, resp)
		})
	}
}
