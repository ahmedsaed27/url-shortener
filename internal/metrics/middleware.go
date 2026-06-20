package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type statusResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}

	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

// HTTPMetricsMiddleware records the request count and duration for API routes.
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(writer, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}

		labels := []string{r.Method, route, strconv.Itoa(writer.status)}
		HTTPRequestsTotal.WithLabelValues(labels...).Inc()
		HTTPRequestDurationSeconds.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
	})
}
