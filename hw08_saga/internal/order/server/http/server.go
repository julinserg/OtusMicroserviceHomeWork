package order_internalhttp

import (
	"context"
	"errors"
	"net/http"

	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

type SrvOrder interface {
	CreateOrder(user order_app.Order) error
	CancelOrder(id string) error
	StatusOrder(id string) (string, error)
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

func NewServer(logger Logger, srvOrder SrvOrder, endpoint string) *Server {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    endpoint,
		Handler: loggingMiddleware(mux, logger),
	}

	uh := ordersHandler{logger: logger, srvOrder: srvOrder}
	mux.HandleFunc("/api/v1/orders/health", hellowHandler)
	mux.HandleFunc("/api/v1/orders/create", uh.createHandler)
	mux.HandleFunc("/api/v1/orders/cancel", uh.cancelHandler)
	mux.HandleFunc("/api/v1/orders/status", uh.statusHandler)
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
