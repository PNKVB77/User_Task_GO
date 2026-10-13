package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/PNKVB77/User_Task_GO/internal/models"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// порядок аргументов регламетирован:
// заголовки ставим до WriteHeader - после него уже все улетели по сети
// WriteHeader - до первого Write, потому что первый сам поставит 200, если статус не выставлен. Баг: Человек пишет тело
// WriteHeader(201) получает 200, и warning в логах
func writeJson(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json, charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("err response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJson(w, status, ErrorResponse{Error: msg})
}

func WriteServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrorNotFound):
		writeError(w, http.StatusNotFound, "tasks not found")
	case errors.Is(err, models.ErrorInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		// сюда попадает из-за ниожиданного - ошибки драйвера, паники в другом вместе
		slog.Error("unhandled service err", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server")
	}

}
