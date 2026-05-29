// Command api is the application entrypoint.
// It wires all dependencies (config, logger, database, adapters, use cases,
// handlers) and starts the HTTP server with the outbox background worker.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	adapterhttp "github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/inbound/http"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/inbound/http/handler"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/model"
	persistence "github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/repository"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/pipefy"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/port/service"
	uclient "github.com/Casagrande-Lucas/pipefy-integration/internal/usecase/client"
	uoutbox "github.com/Casagrande-Lucas/pipefy-integration/internal/usecase/outbox"
	uwebhook "github.com/Casagrande-Lucas/pipefy-integration/internal/usecase/webhook"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/config"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/database"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/logger"
)

const (
	outboxInterval  = 5 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	// ── Config ──────────────────────────────────────────────────────────────
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	// ── Logger ──────────────────────────────────────────────────────────────
	log, err := logger.New(cfg.Logger.Level, cfg.Logger.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting pipefy-integration",
		zap.String("env", string(cfg.App.Environment)),
		zap.Int("port", cfg.App.Port),
	)

	// ── Database ─────────────────────────────────────────────────────────────
	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		log.Fatal("database connect", zap.Error(err))
	}

	if err := database.Migrate(db,
		&model.Client{},
		&model.ProcessedEvent{},
		&model.OutboxEvent{},
	); err != nil {
		log.Fatal("database migrate", zap.Error(err))
	}

	// ── Pipefy adapter ───────────────────────────────────────────────────────
	// dev  → FakeAdapter (no HTTP calls, logs payloads)
	// staging/prod → RealAdapter (simulate=true in staging)
	var pipefySvc service.PipefyService
	if cfg.App.Environment.UsesFakeAdapter() {
		pipefySvc = pipefy.NewFakeAdapter(log)
		log.Info("using fake pipefy adapter")
	} else {
		pipefySvc = pipefy.NewRealAdapter(
			cfg.Pipefy.Token, cfg.Pipefy.PipeID, cfg.Pipefy.Simulate, log,
		)
		log.Info("using real pipefy adapter", zap.Bool("simulate", cfg.Pipefy.Simulate))
	}

	// ── Repositories (non-transactional) ────────────────────────────────────
	// Used by the outbox worker which manages its own retry logic outside a tx.
	clientRepo := persistence.NewClientRepository(db)
	outboxRepo := persistence.NewOutboxRepository(db)

	// ── Use cases ────────────────────────────────────────────────────────────
	// CreateClient: client row + outbox event committed atomically.
	createClientTx := database.NewGORMTransactor(db, func(tx *gorm.DB) uclient.CreateClientRepos {
		return uclient.CreateClientRepos{
			Client: persistence.NewClientRepository(tx),
			Outbox: persistence.NewOutboxRepository(tx),
		}
	})
	createClientUC := uclient.NewCreateClient(createClientTx)

	// ProcessWebhook: status update + idempotency record committed atomically.
	processWebhookTx := database.NewGORMTransactor(db, func(tx *gorm.DB) uwebhook.ProcessWebhookRepos {
		return uwebhook.ProcessWebhookRepos{
			Client:         persistence.NewClientRepository(tx),
			ProcessedEvent: persistence.NewProcessedEventRepository(tx),
		}
	})
	processWebhookUC := uwebhook.NewProcessWebhook(processWebhookTx)

	// ProcessOutbox: non-transactional — polls pending events and dispatches.
	processOutboxUC := uoutbox.NewProcessOutbox(outboxRepo, clientRepo, pipefySvc, log, 0)

	// ── Gin mode ─────────────────────────────────────────────────────────────
	if cfg.App.Environment == config.Prod {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── Router ───────────────────────────────────────────────────────────────
	clientHandler := handler.NewClientHandler(createClientUC, log)
	webhookHandler := handler.NewWebhookHandler(processWebhookUC, log)
	router := adapterhttp.NewRouter(log, clientHandler, webhookHandler)

	// ── Outbox worker ────────────────────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runOutboxWorker(ctx, processOutboxUC, log)

	// ── HTTP server ──────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: router,
	}

	go func() {
		log.Info("server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// ── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown signal received")
	cancel() // stop outbox worker

	shutCtx, shutCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("forced shutdown", zap.Error(err))
	}

	log.Info("server stopped")
}

// runOutboxWorker polls the outbox every outboxInterval until ctx is cancelled.
func runOutboxWorker(ctx context.Context, uc *uoutbox.ProcessOutbox, log *zap.Logger) {
	ticker := time.NewTicker(outboxInterval)
	defer ticker.Stop()

	log.Info("outbox worker started", zap.Duration("interval", outboxInterval))

	for {
		select {
		case <-ticker.C:
			result, err := uc.Execute()
			if err != nil {
				log.Error("outbox worker error", zap.Error(err))
				continue
			}
			if result.Processed > 0 || result.Failed > 0 {
				log.Info("outbox worker run",
					zap.Int("processed", result.Processed),
					zap.Int("failed", result.Failed),
				)
			}
		case <-ctx.Done():
			log.Info("outbox worker stopped")
			return
		}
	}
}
