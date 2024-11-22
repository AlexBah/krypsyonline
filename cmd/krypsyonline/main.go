package main

import (
	"log/slog"
	"net/http"
	"os"

	"main.go/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting application")

	err := listenAndServe(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		panic(err)
	}
}

// listens to a port, choosing between a secure or unsecured connection
func listenAndServe(certFile, keyFile string) error {
	var err error
	if (certFile == "not exist") || (keyFile == "not exist") {
		listenSecondaryPort(":443")
		err = http.ListenAndServe(":80", http.FileServer(http.Dir("./web")))
	} else {
		listenSecondaryPort(":80")
		err = http.ListenAndServeTLS(":443", certFile, keyFile, http.FileServer(http.Dir("./web")))
	}
	return err
}

// listen secondary port and redirect request to primary port
func listenSecondaryPort(port string) {
	go func() {
		err := http.ListenAndServe(port, http.HandlerFunc(redirectHTTPports))
		if err != nil {
			panic(err)
		}
	}()
}

// redirect HTTP and HTTPs
func redirectHTTPports(w http.ResponseWriter, r *http.Request) {
	target := r.Host + r.URL.Path
	if r.TLS != nil {
		target = "http://" + target
	} else {
		target = "https://" + target
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
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
