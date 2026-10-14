package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PNKVB77/User_Task_GO/config"
	"github.com/PNKVB77/User_Task_GO/internal/handlers"
	"github.com/PNKVB77/User_Task_GO/internal/repo"
	"github.com/PNKVB77/User_Task_GO/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}

}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return err
	}
	defer pool.Close()

	// Ping при старте: лучше падать на старте чем принимать трафик чем и 500
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return err
	}
	slog.Info("connected to database")

	taskRepo := repo.NewTaskRepo(pool)
	srvs := services.New(taskRepo)
	handler := handlers.NewTaskHandler(srvs)
	router := handlers.NewRouter(handler, pool)

	// ListenAndServe удобен но не исмеет таймаута
	// Клиент открывает  соедиение,  посылает медленно запрос и не досылает - горутина вестьб вечно. Сотня таких случаеев кладет сервер
	// при нулевой РЕАЛЬНОЙ нагрузке (атака Slowloris)
	srv := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,  // скольк ждет запрос от клиента
		WriteTimeout: cfg.HTTP.WriteTimeout, // скольк времени на ответ
		IdleTimeout:  cfg.HTTP.IdleTimeout,  // сколько держим молчаний "keep-alive"
	}
	// Сервер - в горутину, мэйн блокируется в ожидании сигнала
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server started", "addr", cfg.HTTP.Address)
		// ErrServerClose - Shutdown, а не ошибка
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		return err
	case <-sigCtx.Done():
		slog.Info("shutdown signal received")
	}

	// Graceful Shutdown
	// Shutdown делает две вещи: перестает принимать новые соединения и ждет, пока деработают другие.
	// Дэдлайн обязателен, так как один запрос может повиснуть в проде и не дать ему умереть.

	shutdownCtx, cancelShutDown := context.WithTimeout(ctx, cfg.HTTP.ShutdownTimeout)
	defer cancelShutDown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	slog.Info("server stopped gracefully")

	return nil
}
