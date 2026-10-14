package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/PNKVB77/User_Task_GO/internal/models"
)

// Handler не знает про сервисный слой
type TaskService interface {
	Create(ctx context.Context, in models.CreateTaskRequest) (models.Task, error)
	GetById(ctx context.Context, id int64) (models.Task, error)
	Update(ctx context.Context, id int64, in models.UpdateTaskRequest) (models.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, f models.ListFilter) ([]models.Task, error)
}

type TaskHandler struct {
	svc TaskService
}

func NewTaskHandler(svc TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) PathId(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id must be positive")
		return 0, false
	}
	return id, true
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "ugly JSON body")
		return
	}
	task, err := h.svc.Create(r.Context(), in)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	writeJson(w, http.StatusCreated, task)

}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var filter models.ListFilter
	if raw := query.Get("done"); raw != "" {
		done, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "done must be t/f")
			return
		}
		filter.Done = &done
	}

	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "limit must be a number")
			return
		}
		filter.Limit = limit
	}

	task, err := h.svc.List(r.Context(), filter)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	writeJson(w, http.StatusOK, task)
}

func (h *TaskHandler) GetById(w http.ResponseWriter, r *http.Request) {
	id, ok := h.PathId(w, r)
	if !ok {
		return
	}
	task, err := h.svc.GetById(r.Context(), id)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	writeJson(w, http.StatusOK, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := h.PathId(w, r)
	if !ok {
		return
	}
	var in models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "ugly JSON body")
		return
	}
	task, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	writeJson(w, http.StatusOK, task)

}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.PathId(w, r)
	if !ok {
		return
	}
	err := h.svc.Delete(r.Context(), id)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
