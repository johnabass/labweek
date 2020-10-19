package main

import (
	"net/http"
)

type PluginHandler struct {
	H func(http.ResponseWriter, *http.Request)
}

func (ph PluginHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ph.H(response, request)
	response.Header().Set("Content-Type", "application/json")
	response.Write([]byte(
		`{"plugin": "true"}`,
	))
}
