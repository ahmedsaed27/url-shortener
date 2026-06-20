package app

import (
	"github.com/as9840935/url-shortener/internal/health"
	"github.com/as9840935/url-shortener/internal/shorturl"
)

type Handlers struct {
	Health   *health.Handler
	ShortURL *shorturl.Handler
}

func RegisterHandlers(app *Application) Handlers {
	shortURLCache := shorturl.NewRedisCache(app.Redis)

	shortURLService := shorturl.NewService(
		app.Store.ShortURLs,
		shortURLCache,
		app.Config.AppBaseURL,
		app.Config.ShortCodeLength,
		app.Config.ShortURLCacheTTL,
		app.Config.ShortURLNegativeCacheTTL,
	)

	return Handlers{
		Health:   health.NewHandler(app.DB, app.Redis),
		ShortURL: shorturl.NewHandler(shortURLService, app.Validate, app.AnalyticsProducer),
	}
}
