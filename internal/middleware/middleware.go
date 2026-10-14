package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type Middleware func(http.Handler) http.Handler

type ctxKey struct{ name string }

var requestIdKey = ctxKey{"request_id"}

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func RequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newId()
		}
		w.Header().Set("X-RequestId", id)
		ctx := context.WithValue(r.Context(), requestIdKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func RequestIdForm(ctx context.Context) string {
	id, _ := ctx.Value(requestIdKey).(string)
	return id
}

type StatusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *StatusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &StatusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		slog.Info("request", "method", r.Method, "status", rec.status, "duration", time.Since(start), "request_id", RequestIdForm(r.Context()))
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"err", rec,
					"stack", string(debug.Stack()),
					"request_id", RequestIdForm(r.Context()),
				)

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "internal error"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func newId() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
