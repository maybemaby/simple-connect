package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"simple-connect/api"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

type Args struct {
	Port string
}

func argParse() Args {
	var args Args
	flag.StringVar(&args.Port, "port", "8000", "port to listen on")
	flag.Parse()
	return args
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	location, err := time.LoadLocation("UTC")
	if err != nil {
		log.Println("Error loading location")
	}

	time.Local = location
}

func main() {
	args := argParse()

	ctx, cancel := context.WithCancel(context.Background())

	loadEnv()

	// Server
	appEnv := os.Getenv("APP_ENV")
	allowedHosts := os.Getenv("ALLOWED_HOSTS")

	hosts := strings.Split(allowedHosts, ",")

	isDebug := appEnv == "development"

	var logLevel slog.Level

	if isDebug {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	cfg := api.ServerConfig{
		Port:         args.Port,
		LogLevel:     logLevel,
		AllowedHosts: hosts,
	}

	server, err := api.NewServer(cfg, !isDebug)

	if err != nil {
		log.Fatalf("Error creating server: %v", err)
	}

	defer server.Cleanup(ctx)

	// OS Signals
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		err := server.Start()

		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-osSignals

	cancel()
}
