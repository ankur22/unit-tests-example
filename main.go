package main

import (
	"errors"
	"fmt"
	"strings"
)

// URLShortner will return a shortened url of the input url
// e.g. https://google.com/search?q=blah -> urls.com/12
type URLShortner struct {
	baseURL string
	store   map[string]string
	i       int
}

// Shorten will convert urls to a shortened url
// e.g. https://google.com/search?q=blah -> urls.com/12
func (u *URLShortner) Shorten(in string) (string, error) {
	if !strings.Contains(in, ".com") {
		return "", errors.New("invalid")
	}

	if !strings.Contains(in, "https://") {
		in = fmt.Sprintf("https://%s", in)
	}

	u.i++
	val := fmt.Sprintf("%s/%d", u.baseURL, u.i)
	u.store[val] = in

	return val, nil
}

// Get the full URL with the shortened URL
func (u *URLShortner) Get(in string) (string, error) {
	val, ok := u.store[in]
	if !ok {
		return "", errors.New("Not found")
	}

	return val, nil
}
