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
				},
				{
					"domain": "example.com",
					"path": "/sample2",
					"ready": false
				},
				{
					"domain": "example1.com",
					"path": "/sample1*",
					"ready": true,
					"rule_handlers": [
						{
							"name": "path_replace",
							"config": {
								"search": "/sample1",
								"replace": "",
								"times": 1
							}
						}
					]
				}
			]
		`))
			return
		}

		fmt.Println("Rest of api")
		w.Write([]byte(r.URL.String()))
	}

	err := http.ListenAndServe(":8000", http.HandlerFunc(handler))
	if err != nil {
		panic(err)
	}
}
