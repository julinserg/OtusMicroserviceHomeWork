package order_internalhttp

import (
	"context"
	"errors"
	"net/http"

	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

type SrvOrder interface {
	CreateOrder(user order_app.Order) error
	GetOrdersCount() (int, error)
}

type Storage interface {
	GetOrCreateRequest(id string) (order_app.Request, error)
	UpdateRequest(obj order_app.Request) error
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

func NewServer(logger Logger, storage Storage, srvOrder SrvOrder, endpoint string) *Server {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    endpoint,
		Handler: loggingMiddleware(mux, logger),
	}

	uh := ordersHandler{logger: logger, storage: storage, srvOrder: srvOrder}
	mux.HandleFunc("/api/v1/orders/health", hellowHandler)
	mux.HandleFunc("/api/v1/orders/create", uh.createHandler)
	mux.HandleFunc("/api/v1/orders/count", uh.countHandler)
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
