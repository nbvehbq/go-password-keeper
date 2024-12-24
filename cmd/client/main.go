package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nbvehbq/go-password-keeper/internal/client"
	"github.com/nbvehbq/go-password-keeper/internal/commander"
	"github.com/nbvehbq/go-password-keeper/internal/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"

	resorces = []string{"All", "Login & password", "Text", "Binary (file)", "Bank card"}
)

func main() {
	greetings := "Build version: %s\nBuild date: %s\nBuild commit: %s\n\n"
	fmt.Printf(greetings, buildVersion, buildDate, buildCommit)

	cfg, err := client.NewConfig()
	if err != nil {
		log.Fatal(err, "Load config")
	}

	if errInit := logger.Initialize("info"); errInit != nil {
		log.Fatal(errInit, "initialize logger")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, err := client.NewClient(ctx, cfg)
	if err != nil {
		log.Fatal(err, "create client")
	}

	shell := commander.SetupCommands(ctx, client)
	shell.Run()
}
