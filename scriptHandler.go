package main

import (
	"net/http"

	"github.com/robertkrimen/otto"
)

type ScriptHandler struct {
	S *otto.Script
}

func (sh ScriptHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	vm := otto.New()
	vm.Set("header", response.Header().Add)
	_, err := vm.Run(sh.S)
	if err != nil {
		response.Header().Set("Script-Error", err.Error())
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.Write([]byte(
		`{"script": "true"}`,
	))
}
