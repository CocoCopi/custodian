// Command custodian-worker runs only the background task worker (build +
// deploy execution), for deployments where the API and worker are scaled
// independently.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CocoCopi/custodian/internal/config"
	"github.com/CocoCopi/custodian/internal/jobs"
	"github.com/CocoCopi/custodian/internal/store"
	"github.com/CocoCopi/custodian/internal/ws"
	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("custodian-worker: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("custodian-worker: %v", err)
	}
	defer st.Close()

	hub := ws.NewHub()
	worker := jobs.NewWorker(st, hub, cfg.PublicURL, cfg.Engine, cfg.DeployRoot)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr, Password: cfg.RedisPassword},
		asynq.Config{Concurrency: 4},
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("custodian-worker shutting down...")
		srv.Shutdown()
	}()

	log.Printf("custodian-worker listening on redis %s (engine=%s)", cfg.RedisAddr, cfg.Engine)
	if err := srv.Run(worker.Handler()); err != nil {
		log.Printf("custodian-worker stopped: %v", err)
	}
}
