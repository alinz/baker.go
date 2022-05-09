package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alinz/baker.go/confutil"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	http.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World! service")
	})

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		confutil.NewEndpoints().
			New("example.com", "/", true).
			Done(w)
	})

	fmt.Printf("Starting server at port 8000\n")

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
