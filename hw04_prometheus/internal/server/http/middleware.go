package internalhttp

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests.",
}, []string{"path"})

func initPrometheus() {
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
}

func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		path := r.Method + ":" + r.URL.Path
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))

		recorder := &StatusRecorder{
			ResponseWriter: w,
			Status:         0,
		}

		next.ServeHTTP(recorder, r)
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		var sb strings.Builder
		sb.WriteString(ip + " ")
		sb.WriteString("[" + startTime.String() + "] ")
		sb.WriteString(r.Method + " ")
		sb.WriteString(r.URL.Path + " ")
		sb.WriteString(r.Proto + " ")
		sb.WriteString(strconv.Itoa(recorder.Status) + " ")
		sb.WriteString(time.Since(startTime).String() + " ")
		sb.WriteString(`'` + r.UserAgent() + `'`)
		logger.Info(sb.String())

		responseStatus.WithLabelValues(strconv.Itoa(recorder.Status)).Inc()
		totalRequests.WithLabelValues(path).Inc()
		timer.ObserveDuration()
	})
}
