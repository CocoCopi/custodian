// Command custodian-api runs the Custodian control plane: REST API,
// background worker, and live log streaming.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CocoCopi/custodian/internal/api"
	"github.com/CocoCopi/custodian/internal/auth"
	"github.com/CocoCopi/custodian/internal/config"
	"github.com/CocoCopi/custodian/internal/jobs"
	"github.com/CocoCopi/custodian/internal/store"
	"github.com/CocoCopi/custodian/internal/ws"
	"github.com/hibiken/asynq"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("custodian-api: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	// Live log hub
	hub := ws.NewHub()

	// Auth
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.TokenTTL)
	oidc, err := auth.NewOIDCProvider(ctx, cfg.OIDCIssuer, cfg.OIDCClientID, cfg.OIDCClientSecret, cfg.OIDCRedirectURL)
	if err != nil {
		return err
	}

	// Background worker
	queue := jobs.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	defer func() { _ = queue.Close() }()

	worker := jobs.NewWorker(st, hub, cfg.PublicURL, cfg.Engine, cfg.DeployRoot)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr, Password: cfg.RedisPassword},
		asynq.Config{Concurrency: 4, Logger: asynqLogger{}},
	)
	go func() {
		// Run blocks until Shutdown() is called; any other error is fatal.
		if err := srv.Run(worker.Handler()); err != nil {
			log.Printf("asynq server stopped: %v", err)
		}
	}()
	defer srv.Shutdown()

	// HTTP API
	server := api.New(cfg, st, tokens, oidc, queue, hub)
	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           server.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("custodian-api listening on %s (engine=%s)", cfg.Addr, cfg.Engine)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	return httpSrv.Shutdown(shutdownCtx)
}

type asynqLogger struct{}

func (asynqLogger) Debug(args ...any) { log.Println(append([]any{"[asynq:debug]"}, args...)...) }
func (asynqLogger) Info(args ...any)  { log.Println(append([]any{"[asynq:info]"}, args...)...) }
func (asynqLogger) Warn(args ...any)  { log.Println(append([]any{"[asynq:warn]"}, args...)...) }
func (asynqLogger) Error(args ...any) { log.Println(append([]any{"[asynq:error]"}, args...)...) }
func (asynqLogger) Fatal(args ...any) { log.Fatal(append([]any{"[asynq:fatal]"}, args...)...) }
