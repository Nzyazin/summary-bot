package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nzyazin/summary-bot/internal/app"
	"github.com/nzyazin/summary-bot/internal/app/config"
)

func main() {
	cfg, err := config.LoadConfig("config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	go func() {
		if err := application.Run(ctx); err != nil {
			log.Printf("Error running application: %v", err)
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down application...")
	if err := application.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Application stopped")
}
