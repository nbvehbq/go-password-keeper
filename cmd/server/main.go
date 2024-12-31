package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nbvehbq/go-password-keeper/internal/logger"
	"github.com/nbvehbq/go-password-keeper/internal/server"
	"github.com/nbvehbq/go-password-keeper/internal/session"
	"github.com/nbvehbq/go-password-keeper/internal/storage/postgres"
	"golang.org/x/sync/errgroup"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	greetings := "Build version: %s\nBuild date: %s\nBuild commit: %s\n\n"
	fmt.Printf(greetings, buildVersion, buildDate, buildCommit)

	cfg, err := server.NewConfig()
	if err != nil {
		log.Fatal(err, "Load config")
	}

	if errInit := logger.Initialize(cfg.LogLevel); errInit != nil {
		log.Fatal(errInit, "initialize logger")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner, ctx := errgroup.WithContext(ctx)

	session := session.NewSessionStorage(ctx)
	storage, err := postgres.NewStorage(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err, "create storage")
	}

	server, err := server.NewServer(storage, session, cfg)
	if err != nil {
		log.Fatal(err, "create server")
	}

	runner.Go(func() error {
		return server.Run()
	})

	runner.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(ctx)
	})

	if err := runner.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}
}
