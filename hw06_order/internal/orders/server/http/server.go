package order_internalhttp

import (
	"context"
	"errors"
	"net/http"

	orders_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw06_order/internal/orders/app"
)

type Storage interface {
	CreateOrder(user orders_app.Order) error
	GetRequest(id string) (orders_app.Request, error)
	SaveRequest(obj orders_app.Request) error
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

	uh := ordersHandler{logger, storage}
	mux.HandleFunc("/api/v1/orders/health", hellowHandler)
	mux.HandleFunc("/api/v1/orders/create", uh.commonHandler)
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
