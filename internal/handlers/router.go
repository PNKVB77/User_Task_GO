package handlers

import (
	"context"
	"net/http"

	"github.com/PNKVB77/User_Task_GO/internal/middleware"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func NewRouter(h *TaskHandler, db Pinger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/tasks", h.Create)
	mux.HandleFunc("GET /api/v1/tasks", h.List)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.GetById)
	mux.HandleFunc("PATCH /api/v1/tasks/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.Delete)

	// healthz вне api/v1 - это не часть публичного контракта, а метод для орекстровки и мониторинга
	mux.HandleFunc("GET /healthz", healthz(db))

	return middleware.Chain(mux, middleware.RequestId, middleware.Logger, middleware.Recovery)

}

func healthz(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavaliable")
			return
		}
		writeJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
