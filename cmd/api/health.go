package main

import (
	"net/http"
)

func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	// Return a 200 OK response with a simple message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}