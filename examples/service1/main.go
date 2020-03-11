package main

import (
	"fmt"
	"net/http"
)

func main() {

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Got a request for ", r.URL.Path)

		if r.URL.Path == "/config" {
			fmt.Println("sending back config")
			w.Write([]byte(`
			[
				{
					"domain": "example.com",
					"path": "/sample1",
					"ready": true
				}
			]
		`))
			return
		}

		fmt.Println("Rest of api")
		w.Write([]byte(r.URL.Path))
	}

	err := http.ListenAndServe(":8000", http.HandlerFunc(handler))
	if err != nil {
		panic(err)
	}
}