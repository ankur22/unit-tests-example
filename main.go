package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	u := &URLShortner{
		i:     0,
		store: make(map[string]string),
	}

	c := make(chan string)
	o := make(chan string)
	defer close(c)
	defer close(o)
	go cliInput(ctx, c, o)
	go cliFronted(ctx, u, c, o)

	go httpFronted(ctx, u)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)

	done := make(chan bool, 1)
	go func() {
		<-sig
		fmt.Println("Shutting down service")
		done <- true
	}()

	fmt.Println("Started")
	<-done
	cancel()
	fmt.Println("Exiting")
}

func cliInput(ctx context.Context, c chan<- string, o <-chan string) {
	go func() {
		for {
			var in string
			_, err := fmt.Scanln(&in)
			if err != nil {
				fmt.Printf("Caught error %v\n", err)
				continue
			}
			c <- in
			out := <-o
			fmt.Printf("Shotened URL: %s\n", out)
		}
	}()

	<-ctx.Done()
	fmt.Printf("Close cliInput\n")
}

func cliFronted(ctx context.Context, u *URLShortner, c <-chan string, o chan<- string) {
	go func() {
	outer:
		for {
			select {
			case in := <-c:
				resp, err := u.Shorten(in)
				if err != nil {
					fmt.Printf("Caught error %v\n", err)
					break
				}
				o <- resp
			case <-ctx.Done():
				fmt.Printf("Close cliFrontend\n")
				break outer
			}
		}
	}()
}

type body struct {
	Index string
}

func httpFronted(ctx context.Context, u *URLShortner) {
	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		r := mux.NewRouter()

		r.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
			url := r.FormValue("u")

			resp, err := u.Shorten(url)
			if err != nil {
				w.WriteHeader(400)
				w.Write([]byte(fmt.Sprintf("Not a valid URL '%s'\n", url)))
				return
			}

			b, err := json.Marshal(body{Index: resp})
			if err != nil {
				fmt.Printf("Error while marshalling json %v\n", err)
				w.WriteHeader(500)
				return
			}

			w.WriteHeader(201)
			w.Header().Add("content-type", "application/json")
			w.Write(b)
			return
		}).Queries("u", "{url}").Methods("POST")

		r.HandleFunc("/{i}", func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			i := vars["i"]

			resp, err := u.GetURL(i)
			if err != nil {
				w.WriteHeader(404)
				w.Write([]byte(fmt.Sprintf("Not found '%s'\n", i)))
				return
			}

			http.Redirect(w, r, resp, 302)
			return
		})

		s.Handler = r

		fmt.Println("Starting server")
		err := s.ListenAndServe()
		if err != nil {
			fmt.Printf("Error while trying to start server %v\n", err)
		}

		fmt.Println("Server shutting down")
	}()

	<-ctx.Done()
	fmt.Println("Shut down server")
	s.Shutdown(ctx)
	fmt.Printf("Close httpFrontend\n")
}

// URLShortner can create ids for the original
// URLs and return the original URLs with the id
type URLShortner struct {
	m     sync.Mutex
	store map[string]string
	i     int
}

// Shorten will convert urls to a id
// e.g. https://google.com/search?q=blah -> 12
func (u *URLShortner) Shorten(in string) (string, error) {
	_, err := url.ParseRequestURI(in)
	if err != nil {
		return "", errors.New("invalid")
	}

	if !strings.Contains(in, ".") {
		return "", errors.New("invalid")
	}

	var val string
	u.m.Lock()
	u.i++
	val = strconv.Itoa(u.i)
	u.store[val] = in
	u.m.Unlock()

	return val, nil
}

// GetURL the full URL with the id
// e.g. 12 -> https://google.com/search?q=blah
func (u *URLShortner) GetURL(in string) (string, error) {
	val, ok := u.store[in]
	if !ok {
		return "", errors.New("Not found")
	}

	return val, nil
}
