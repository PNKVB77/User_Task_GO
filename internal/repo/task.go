package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/PNKVB77/User_Task_GO/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{pool: pool}
}

const TaskColumns = `id, title, done, due_date, created_at`

const query = `
	INSERT INTO tasks (title, due_date)
	VALUES ($1, $2)
	RETURNING ` + TaskColumns

// $1,$2 - параметры. Драйвер отправляет запрос и значение отдельно, поэтому постгрес не может распознать значение как SQL
// RETERNING нужен чтобы забрать сгенированный id и created_at без SELECT

func (r *TaskRepo) Create(ctx context.Context, in models.CreateTaskRequest) (models.Task, error) {
	var t models.Task
	err := r.pool.QueryRow(ctx, query, in.Title, in.DueDate).Scan(&t.ID, &t.Title, &t.Done, &t.DueDate, &t.CreatedAt)
	if err != nil {
		return models.Task{}, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

func (r *TaskRepo) GetById(ctx context.Context, id int64) (models.Task, error) {
	const query = `SELECT ` + TaskColumns + ` FROM tasks WHERE id = $1`
	var t models.Task
	err := r.pool.QueryRow(ctx, query, id).Scan(&t.ID, &t.Title, &t.Done, &t.DueDate, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Task{}, models.ErrorNotFound
	}
	if err != nil {
		return models.Task{}, fmt.Errorf("get task %d: %w", id, err)
	}
	return t, nil
}

// errors.Is а не == - если обернуть ошибку через %w сравнение на равенство перестанет работать, а Is развернет цепочку.

func (r *TaskRepo) List(ctx context.Context, f models.ListFilter) ([]models.Task, error) {
	const query = `SELECT ` + TaskColumns + ` FROM tasks WHERE ($1::bool IS NULL OR done = $1)
	ORDER BY created_at DESC, id DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, f.Done, f.Limit)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.DueDate, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tasks: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)

	}
	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, id int64, in models.UpdateTaskRequest) (models.Task, error) {
	const query = `UPDATE tasks 
	SET title = COALESCE($2, title),
		done = COALESCE ($3, done),
		due_date = COALESCE ($4, due_date)
	WHERE id = $1
	RETURNING ` + TaskColumns
	var t models.Task
	err := r.pool.QueryRow(ctx, query, id, in.Title, in.Done, in.DueDate).Scan(&t.ID, &t.Title, &t.Done, &t.DueDate, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Task{}, models.ErrorNotFound
	}

	if err != nil {
		return models.Task{}, fmt.Errorf("update task %d: %w", id, err)
	}
	return t, nil
}

func (r *TaskRepo) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task %d: %w", id, err)
	}
	// удаление строки которой нет - не ошибка на уровне SQL: ЗАПРОС УСПЕШЕН просто нет удаления
	if tag.RowsAffected() == 0 {
		return models.ErrorNotFound
	}
	return nil
}
