package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/shalimski/dummyhttp/config"
	"github.com/shalimski/dummyhttp/handlers"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("loading config", "error", err.Error())
		os.Exit(1)
	}

	handler, err := handlers.New(cfg.Mode, cfg.Handler)
	if err != nil {
		slog.Error("creating handler", "error", err.Error())
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv := http.Server{
		Addr:              cfg.Server.Listen,
		Handler:           handler,
		ReadHeaderTimeout: cfg.Server.Timeout,
		WriteTimeout:      cfg.Server.Timeout,
		IdleTimeout:       cfg.Server.Timeout,
	}
	slog.Info("starting server", "listen", cfg.Server.Listen)

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("shutdown", "error", err.Error())
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
		} else {
			slog.Info("server stopped")
		}
	}
}
