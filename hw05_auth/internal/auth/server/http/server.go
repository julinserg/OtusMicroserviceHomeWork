package auth_internalhttp

import (
	"context"
	"errors"
	"net/http"

	auth_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw05_auth/internal/auth/app"
)

type Storage interface {
	RegisterUser(user auth_app.UserAuth) (int, error)
	GetUser(login string, password string) (auth_app.UserAuth, error)
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
	sessions := make(map[string]auth_app.UserAuth)
	uh := userHandler{logger, storage, sessions}
	mux.HandleFunc("/health", hellowHandler)
	mux.HandleFunc("/register", uh.registerHandler)
	mux.HandleFunc("/login", uh.loginHandler)
	mux.HandleFunc("/logout", uh.logoutHandler)
	mux.HandleFunc("/auth", uh.authHandler)
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
