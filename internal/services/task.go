package services

import (
	"context"

	"github.com/PNKVB77/User_Task_GO/internal/models"
)

// TaskRepository находится здесь, в пакете потрибителя, а не в пакете repo
// Это идиома в GO: интерфейс пренадлежить тому, кто его использует
// 1. Реализаций может быть несколько
// 2. Здесь только методы для сервиса, а не всего репозитория

type TaskRepository interface {
	Create(ctx context.Context, in models.CreateTaskRequest) (models.Task, error)
	GetById(ctx context.Context, id int64) (models.Task, error)
	Update(ctx context.Context, id int64, in models.UpdateTaskRequest) (models.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, f models.ListFilter) ([]models.Task, error)
}

const (
	defaultLimit    = 20
	maxAllowedLimit = 100
)

type TaskService struct {
	repo TaskRepository
}

func NewTaskRepository()
