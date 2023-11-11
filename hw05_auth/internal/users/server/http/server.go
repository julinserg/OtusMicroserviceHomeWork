package users_internalhttp

import (
	"context"
	"errors"
	"net/http"

	users_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw05_auth/internal/users/app"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Storage interface {
	CreateUser(user users_app.User) error
	FindUserById(id string) (users_app.User, error)
	UpdateUser(user users_app.User) error
}

type Server struct {
	server   *http.Server
	logger   Logger
	endpoint string
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func NewServer(logger Logger, storage Storage, endpoint string) *Server {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    endpoint,
		Handler: loggingMiddleware(mux, logger),
	}
	initPrometheus()
	uh := userHandler{logger, storage}
	mux.HandleFunc("/users/health", hellowHandler)
	mux.HandleFunc("/users/me", uh.commonHandler)
	mux.Handle("/metrics", promhttp.Handler())
	return &Server{server, logger, endpoint}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("http server started on " + s.endpoint)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	s.logger.Info("http server stopped")
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
