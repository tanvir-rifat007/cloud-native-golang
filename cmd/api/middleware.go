package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// MetricsMiddleware returns an HTTP middleware that records Prometheus metrics.
// This is a constructor that returns a middleware, like a factory.
func (app *application) Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, 1)
			start := time.Now()

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			code := strconv.Itoa(status)

			app.requests.WithLabelValues(r.Method, r.URL.Path, code).Inc()
			app.requestDurations.WithLabelValues(code).Observe(duration.Seconds())
		})
	}
}


func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error("panic recovered:", zap.Error(err.(error)))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}


// for admin 
// func (app *application) requireAdmin(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		 user,_  := app.models.NewsletterSubscribers.GetAdminSubscriber()

//    fmt.Println("Admin??",user.IsAdmin)


// 		if !user.IsAdmin {
// 			app.notFoundResponse(w, r)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

