package main

import (
	"log/slog"
	"os"

	"main.go/internal/config"
	"main.go/internal/handler"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting application")

	shutdownCh := make(chan struct{})
	handler.ListenPortal(cfg.CertFile, cfg.KeyFile, shutdownCh, log)
	handler.ListenStopSig()
	close(shutdownCh)

}

// setup level of logger info
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
