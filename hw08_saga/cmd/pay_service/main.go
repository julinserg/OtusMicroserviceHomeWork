package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/logger"
	pay_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/pay/app"
	pay_internalhttp "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/pay/server/http"
	pay_sqlstorage "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/pay/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/auth_config.toml", "Path to configuration file")
}

/*
 goose -dir migrations postgres "user=sergey password=sergey dbname=calendar sslmode=disable" up
*/

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := NewConfig()
	err := config.Read(configFile)
	if err != nil {
		var value string
		value, _ = os.LookupEnv("USC_LOG_LEVEL")
		config.Logger.Level = value
		value, _ = os.LookupEnv("USC_HTTP_HOST")
		config.HTTP.Host = value
		value, _ = os.LookupEnv("USC_HTTP_PORT")
		config.HTTP.Port = value
		value, _ = os.LookupEnv("USC_HTTP_PORT")
		config.HTTP.Port = value
		dbHost, _ := os.LookupEnv("USC_PG_HOST")
		dbUser, _ := os.LookupEnv("USC_PG_USER")
		dbPassword, _ := os.LookupEnv("USC_PG_PASSWORD")
		dbName, _ := os.LookupEnv("USC_PG_DB")
		config.PSQL.DSN = fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbUser, dbPassword, dbName)
	}

	f, err := os.OpenFile("auth_service_logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalln("error opening file: " + err.Error())
	}
	defer f.Close()

	logg := logger.New(config.Logger.Level, f)

	var storage pay_app.Storage
	{
		sqlstor := pay_sqlstorage.New()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := sqlstor.Connect(ctx, config.PSQL.DSN); err != nil {
			logg.Error("cannot connect to psql: " + err.Error())
		}
		if err := sqlstor.CreateSchema(); err != nil {
			logg.Error("cannot create database schema: " + err.Error())
		}
		defer func() {
			if err := sqlstor.Close(); err != nil {
				logg.Error("cannot close psql connection: " + err.Error())
			}
		}()
		storage = sqlstor
	}

	endpointHttp := net.JoinHostPort(config.HTTP.Host, config.HTTP.Port)
	serverHttp := pay_internalhttp.NewServer(logg, storage, endpointHttp)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := serverHttp.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("auth_service is running...")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serverHttp.Start(ctx); err != nil {
			logg.Error("failed to start http server: " + err.Error())
			cancel()
			return
		}
	}()
	wg.Wait()
}
