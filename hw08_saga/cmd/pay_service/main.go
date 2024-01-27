package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/logger"
	pay_amqp "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/pay/amqp"
	pay_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/pay/app"
	pay_sqlstorage "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/pay/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/pay_config.toml", "Path to configuration file")
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
		dbHost, _ := os.LookupEnv("USC_PG_HOST")
		dbUser, _ := os.LookupEnv("USC_PG_USER")
		dbPassword, _ := os.LookupEnv("USC_PG_PASSWORD")
		dbName, _ := os.LookupEnv("USC_PG_DB")
		config.PSQL.DSN = fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbUser, dbPassword, dbName)
		value, _ = os.LookupEnv("USC_AMQP_URI")
		config.AMQP.URI = value
	}

	f, err := os.OpenFile("pay_service_logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
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

	payMQ := pay_amqp.New(logg, config.AMQP.URI)

	srvPay := pay_app.New(logg, storage, payMQ)

	payMQ.SetService(srvPay)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg.Info("pay_service is running...")

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := payMQ.StartReceiveOrder(ctx); err != nil {
			logg.Error("failed to start MQ worker(order): " + err.Error())
			cancel()
			return
		}
	}()
	go func() {
		defer wg.Done()
		if err := payMQ.StartReceiveStatus(ctx); err != nil {
			logg.Error("failed to start MQ worker(status): " + err.Error())
			cancel()
			return
		}
	}()
	wg.Wait()
}
