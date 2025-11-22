package main

import (
	"ReilBleem13/pull_requests_service/internal/config"
	"ReilBleem13/pull_requests_service/internal/handler"
	"ReilBleem13/pull_requests_service/internal/repository"
	"ReilBleem13/pull_requests_service/internal/repository/database"
	"ReilBleem13/pull_requests_service/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/theartofdevel/logging"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := config.MustGet()

	level := "info"
	if cfg.App.Mode == "debug" {
		level = "debug"
	}

	logger := logging.NewLogger(
		logging.WithIsJSON(level != "debug"),
		logging.WithAddSource(level != "debug"),
		logging.WithLevel(level),
	)

	ctx = logging.ContextWithLogger(ctx, logger)

	db, err := database.NewPostgresDB(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatal("unable to create database connection")
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db.Client())
	teamRepo := repository.NewTeamRepository(db.Client())
	prRepo := repository.NewPullRequestRepository(db.Client())

	svc := service.NewService(userRepo, teamRepo, prRepo, logger)

	httpMux := handler.NewRouter(svc, logger)
	httpAddr := ":" + cfg.App.Port
	httpServer := handler.NewServer(httpAddr, httpMux)

	httpErrCh := make(chan error)

	go func() {
		logging.L(ctx).Info("starting http server", logging.StringAttr("addr", httpAddr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpErrCh <- err
		}
	}()

	logging.WithAttrs(ctx,
		logging.StringAttr("Port", cfg.App.Port),
		logging.StringAttr("Mode", cfg.App.Mode),
		logging.StringAttr("DB_Host", cfg.Database.Host),
		logging.StringAttr("DB_Port", cfg.Database.Port),
		logging.StringAttr("DB_User", cfg.Database.User),
		logging.StringAttr("DB_Name", cfg.Database.DBName),
		logging.StringAttr("DB_Password", cfg.Database.Password),
		logging.StringAttr("DB_Ssslmode", cfg.Database.SSLMode),
	).Info("pr service started working")

	select {
	case <-ctx.Done():
		logging.L(ctx).Info("received shutdown signal, starting graceful shutdown...")
	case err := <-httpErrCh:
		logging.L(ctx).Error("http server failed", logging.ErrAttr(err))
		return
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logging.L(ctx).Error("http server forced shutdown", logging.ErrAttr(err))
	}

	if err := db.Close(); err != nil {
		logging.L(ctx).Error("failed to close database connection", logging.ErrAttr(err))
	}

	<-shutdownCtx.Done()
	if shutdownCtx.Err() == context.DeadlineExceeded {
		logging.L(ctx).Warn("graceful shutdown timed out")
	} else {
		logging.L(ctx).Info("graceful shutdown completed...")
	}
}
