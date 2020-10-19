package main

import "net/http"

func Handle(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Plugin", "Hello, world!")
}
