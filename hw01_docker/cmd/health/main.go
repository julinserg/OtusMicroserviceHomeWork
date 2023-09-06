package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.toml", "Path to configuration file")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"OK\"}"))
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := NewConfig()
	err := config.Read(configFile)
	if err != nil {
		log.Fatalln("failed to read config: " + err.Error())
	}
	endpointHttp := net.JoinHostPort(config.HTTP.Host, config.HTTP.Port)
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    endpointHttp,
		Handler: mux,
	}
	mux.HandleFunc("/health/", healthHandler)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("failed to stop http server: " + err.Error())
		}
	}()

	log.Println("health service is running...")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Println("start http server on " + endpointHttp)
		if err := server.ListenAndServe(); err != nil {
			log.Println("failed to start http server: " + err.Error())
			cancel()
			return
		}
	}()
	wg.Wait()
}
