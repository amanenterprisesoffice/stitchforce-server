package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stitchforce-server/config"
	"stitchforce-server/internal/storage/s3"
	"stitchforce-server/pkg/email/resend"
	"stitchforce-server/pkg/jwt"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/postgres"
)

func Run() {
	cfg := config.Load()
	log := logger.New(cfg.Log.Level)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		log.Error("postgres connection failed", "error", err)
		return
	}
	defer db.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()

	if err := db.DB.PingContext(pingCtx); err != nil {
		log.Error("postgres ping failed", "error", err)
		return
	}

	if err := postgres.Migrate(db.DB, "./migrations"); err != nil {
		log.Error("database migration failed", "error", err)
		return
	}

	storage, err := s3.New(cfg.Storage)
	if err != nil {
		log.Error("failed to initialize storage", "error", err)
		return
	}

	emailer, err := resend.New(cfg.Resend)
	if err != nil {
		log.Fatal("failed to initialize email service", "error", err)
		return
	}

	repositories := NewRepositories(db)
	useCases := NewUseCases(repositories, log, storage, emailer, cfg)

	jwtManager := jwt.New(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.TTL,
	)

	server := NewServer(
		cfg,
		&useCases,
		log,
		jwtManager,
	)

	go func() {
		if err := server.Run(); err != nil &&
			!errors.Is(err, context.Canceled) {
			log.Error("http server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.App.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown failed", "error", err)
	}
}
