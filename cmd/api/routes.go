package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() chi.Router {
	mux:= chi.NewRouter()


mux.Handle("/metrics", app.recoverPanic(promhttp.HandlerFor(app.registry, promhttp.HandlerOpts{})))


  mux.Handle("/activated", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.catchAllClientRequestHandler))))


  mux.Handle("/newsletters", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.catchAllClientRequestHandler))))

mux.Handle("/newsletters/*", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.catchAllClientRequestHandler))))


	 mux.Handle("/*", app.recoverPanic(app.Metrics()(http.StripPrefix("/", http.FileServer(http.Dir("./public"))))))
	 
mux.Get("/api/v1/health", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.healthCheck))).ServeHTTP)

mux.Post("/api/v1/newsletter", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.NewsletterSignup))).ServeHTTP)

mux.Get("/api/v1/newsletter/confirmation", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.NewsletterConfirmation))).ServeHTTP)

mux.Post("/api/v1/newsletter/create", app.recoverPanic(app.Metrics()(app.requireAdmin(http.HandlerFunc(app.createNewsletterHandler)))).ServeHTTP)

mux.Get("/api/v1/newsletters", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.getNewslettersHandler))).ServeHTTP)

mux.Get("/api/v1/newsletter/{id}", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.getNewletterByIdHandler))).ServeHTTP)

// search 
mux.Get("/api/v1/newsletters/search", app.recoverPanic(app.Metrics()(http.HandlerFunc(app.searchNewsletterHandler))).ServeHTTP)

	return mux
}

