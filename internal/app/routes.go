package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Application) MountRoutes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/health", app.Handlers.Health.Check)
	router.Get("/{code}", app.Handlers.ShortURL.Redirect)

	router.Route("/api", func(r chi.Router) {
		r.Post("/urls", app.Handlers.ShortURL.Create)
	})

	return router
}
