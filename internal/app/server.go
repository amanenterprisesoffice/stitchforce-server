package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"stitchforce-server/config"
	httpv1 "stitchforce-server/internal/controller/http/v1"
	"stitchforce-server/pkg/jwt"
	"stitchforce-server/pkg/logger"

	"github.com/go-playground/validator/v10"
)

type Server struct {
	http *http.Server
}

func NewServer(
	cfg config.Config,
	useCases *UseCases,
	log *logger.Logger,
	jwtManager *jwt.Manager,
) *Server {
	router := httpv1.NewRouter(
		useCases.Users,
		useCases.GarmentWorkers,
		log,
		jwtManager,
		validator.New(),
	)

	return &Server{
		http: &http.Server{
			Addr:              ":" + cfg.App.Port,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s *Server) Run() error {
	if err := s.http.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
