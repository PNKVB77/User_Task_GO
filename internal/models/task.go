package models

import (
	"time"
)

type Task struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Done      bool       `json:"done"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateTaskRequest struct {
	Title   string     `json:"title" validate:"required,min=3,max=100"` // теги validate читают в c/c min/max - считает в рунах, не в байтах
	DueDate *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title   *string    `json:"title" validate:"omitempty,min=3,max=100"`
	Done    *bool      `json:"done"`
	DueDate *time.Time `json:"due_date"`
}

// все поля с указателем чтобы не было zero value и мы могли отличить false от "клинет не заполнил поле"

type ListFilter struct {
	Done  *bool // nil - без фильтра по статусу
	Limit int
}
