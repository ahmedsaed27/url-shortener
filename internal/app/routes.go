package app

import (
	"net/http"
	"github.com/as9840935/url-shortener/internal/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *Application) MountRoutes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Handle("/metrics", promhttp.Handler())

	router.Group(func(r chi.Router) {
		r.Use(metrics.HTTPMetricsMiddleware)

		r.Get("/health", app.Handlers.Health.Check)
		r.Get("/{code}", app.Handlers.ShortURL.Redirect)

		r.Route("/api", func(r chi.Router) {
			r.Post("/urls", app.Handlers.ShortURL.Create)
		})
	})

	return router
}
