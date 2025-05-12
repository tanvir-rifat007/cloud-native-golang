package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/matryer/is"
)

func TestHealth(t *testing.T){
 
	t.Run("returns 200", func(t *testing.T){
		is:= is.New(t)

		app:= &application{}

		mux:= chi.NewRouter()

		mux.Get("/api/v1/health", app.healthCheck)

		code, _, _ := makeGetRequest(mux, "/api/v1/health")

		is.Equal(http.StatusOK, code)
	})




	
}


// makeGetRequest and returns the status code, response headers, and the body.

func makeGetRequest(handler chi.Router, target string) (int, http.Header, string) {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	result := res.Result()
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	return result.StatusCode, result.Header, string(bodyBytes)
}

