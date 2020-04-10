package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

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
		t.Run("GetURL"+test.name, func(t *testing.T) {
			s := &URLShortner{i: 0, store: make(map[string]string)}

			s.Shorten(test.in1)
			resp, err := s.GetURL(test.in2)

			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expected, resp)
		})
	}

	u := URLShortner{store: make(map[string]string), i: 0}

	t.Run("Race", func(t *testing.T) {
		l := 1000
		var wg sync.WaitGroup
		for i := 0; i < l; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := u.Shorten("google.com")
				assert.NoError(t, err)
			}()
		}
		wg.Wait()

		resp, err := u.Shorten("google.com")
		assert.NoError(t, err)
		assert.Equal(t, strconv.Itoa(l+1), resp)
	})
}

func BenchmarkURLShortner(b *testing.B) {
	u := URLShortner{store: make(map[string]string), i: 0}

	b.Run("Race", func(b *testing.B) {
		l := 1000
		var wg sync.WaitGroup
		for i := 0; i < l; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := u.Shorten("google.com")
				assert.NoError(b, err)
			}()
		}
		wg.Wait()
	})
}

func TestCLI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	u := &URLShortner{i: 0, store: make(map[string]string)}

	o := make(chan string)
	c := make(chan string)
	defer close(c)
	defer close(o)
	go cliFronted(ctx, u, c, o)

	c <- "google.com"
	out := <-o

	actual, err := u.GetURL(out)
	assert.Equal(t, "https://google.com", actual)
	assert.NoError(t, err)

	actual, err = u.GetURL("foo")
	assert.Error(t, err)

	cancel()
	time.Sleep(1 * time.Second)
}

func TestHTTP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	u := &URLShortner{i: 0, store: make(map[string]string)}

	go httpFronted(ctx, u)

	c := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := c.Get("http://localhost:8080/1")
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	resp, err = c.Get("http://localhost:8080/shorten?u=http:google.com")
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	body := &body{}
	err = json.Unmarshal(b, body)
	assert.NoError(t, err)

	assert.Equal(t, "1", body.Index)

	cancel()
}
