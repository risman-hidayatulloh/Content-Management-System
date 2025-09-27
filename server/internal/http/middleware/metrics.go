package middleware

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var reqDur = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "cms",
		Name:      "http_request_seconds",
		Help:      "request duration",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path", "code"},
)

func init() { prometheus.MustRegister(reqDur) }

func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, code: 200}
			next.ServeHTTP(ww, r)
			reqDur.WithLabelValues(r.Method, r.URL.Path, http.StatusText(ww.code)).Observe(time.Since(start).Seconds())
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(status int) {
	w.code = status
	w.ResponseWriter.WriteHeader(status)
}
