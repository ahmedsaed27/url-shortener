package app

import (
	httpmiddleware "github.com/as9840935/url-shortener/internal/http/middleware"
	"github.com/as9840935/url-shortener/internal/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
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

		r.With(httpmiddleware.RateLimit(app.RateLimiter, httpmiddleware.RateLimitConfig{
			Type:     "resolve",
			Limit:    app.Config.RateLimitResolveLimit,
			Window:   app.Config.RateLimitResolveWindow,
			FailOpen: app.Config.RateLimitFailOpen,
		})).Get("/{code}", app.Handlers.ShortURL.Redirect)

		r.Route("/api", func(r chi.Router) {
			r.With(httpmiddleware.RateLimit(app.RateLimiter, httpmiddleware.RateLimitConfig{
				Type:     "create",
				Limit:    app.Config.RateLimitCreateLimit,
				Window:   app.Config.RateLimitCreateWindow,
				FailOpen: app.Config.RateLimitFailOpen,
			})).Post("/urls", app.Handlers.ShortURL.Create)
		})
	})

	return router
}
