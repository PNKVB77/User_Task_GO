package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PNKVB77/User_Task_GO/internal/models"
	"github.com/go-playground/validator"
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
	repo     TaskRepository
	validate *validator.Validate
}

func NewTaskRepository(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo, validate: validator.New()}
}

func (s *TaskService) Check(v any) error {
	err := s.validate.Struct(v)
	if err == nil {
		return nil
	}
	var vers validator.ValidationErrors
	if errors.As(err, &vers) && len(vers) > 0 {
		// берем первую ошибку: клиенту достаточно одной внятной причины
		return fmt.Errorf("%w: %s", models.ErrorInvalidInput, humanize(vers[0]))
	}
	// сюда попадаем если валидатору отправили не структуру - баг
	return fmt.Errorf("validate: %w", err)
}

// по умолчанию вададатор возвращает ошибки в нечитабильным для клиенте виде - этого показывать нельзя
func humanize(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return "required " + field
	case "min":
		fmt.Printf("%s must be at least %s", field, fe.Param())
	case "max":
		fmt.Printf("%s must be at most %s", field, fe.Param())
	default:
		return field + " invalid"
	}
	return field
}

func (s *TaskService) Create(ctx context.Context, in models.CreateTaskRequest) (models.Task, error) {
	in.Title = strings.TrimSpace(in.Title)
	if err := s.Check(in); err != nil {
		return models.Task{}, err
	}
	return models.Task{}, nil
}

func (s *TaskService) GetById(ctx context.Context, id int64) (models.Task, error) {
	return s.repo.GetById(ctx, id)
}

func (s *TaskService) List(ctx context.Context, f models.ListFilter) ([]models.Task, error) {
	if f.Limit <= 0 {
		f.Limit = defaultLimit
	}
	if f.Limit > maxAllowedLimit {
		f.Limit = maxAllowedLimit
	}
	return s.repo.List(ctx, f)
}

func (s *TaskService) Update(ctx context.Context, id int64, in models.UpdateTaskRequest) (models.Task, error) {
	if in.Title != nil {
		trimmed := strings.TrimSpace(*in.Title)
		in.Title = &trimmed
	}
	if err := s.Check(in); err != nil {
		return models.Task{}, err
	}
	return s.repo.Update(ctx, id, in)
}

func (s *TaskService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
