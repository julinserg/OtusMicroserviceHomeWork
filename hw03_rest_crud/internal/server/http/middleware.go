package internalhttp

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

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
	})
}
