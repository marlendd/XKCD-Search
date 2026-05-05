package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"yadro.com/course/frontend/adapters/api"
	"yadro.com/course/frontend/adapters/web"
	"yadro.com/course/frontend/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	apiClient, err := api.NewClient(cfg.BaseURL, cfg.Timeout, log)
	if err != nil {
		log.Error("cannot init update adapter", "error", err)
		os.Exit(1)
	}

	handler := web.NewHandler(apiClient, log, cfg.TokenTTL)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handler.SearchPage)
	mux.HandleFunc("POST /", handler.Search)
	mux.HandleFunc("GET /login", handler.LoginPage)
	mux.HandleFunc("POST /login", handler.Login)
	mux.Handle("GET /admin", web.RequireAuth(http.HandlerFunc(handler.AdminPage)))
	mux.Handle("POST /admin/update", web.RequireAuth(http.HandlerFunc(handler.Update)))
	mux.Handle("POST /admin/drop", web.RequireAuth(http.HandlerFunc(handler.Drop)))

	server := http.Server{
		Addr:        cfg.Address,
		ReadTimeout: cfg.Timeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
