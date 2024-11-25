package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// listens to a port, choosing between a secure or unsecured connection
func ListenPortal(certFile, keyFile string, shutdownCh <-chan struct{}, log *slog.Logger) {
	if (certFile == "not exist") || (keyFile == "not exist") {
		listenPrimaryPort(":80", certFile, keyFile, shutdownCh, log)
	} else {
		listenPrimaryPort(":443", certFile, keyFile, shutdownCh, log)
		listenSecondaryPort(":80", shutdownCh, log)
	}
}

// listen primary port and return fileserver
func listenPrimaryPort(port, certFile, keyFile string, shutdownCh <-chan struct{}, log *slog.Logger) {
	srv := &http.Server{Addr: port, Handler: http.FileServer(http.Dir("./web"))}
	log.Info(fmt.Sprintf("Starting listen on port %s", srv.Addr))

	if port == ":80" {
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Error("Port", srv.Addr, err)
			}
		}()
	}
	if port == ":443" {
		go func() {
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
				log.Error("Port", srv.Addr, err)
			}
		}()
	}

	stopListen(srv, shutdownCh, log)
}

// listen secondary port and redirect request to HTTPs
func listenSecondaryPort(port string, shutdownCh <-chan struct{}, log *slog.Logger) {
	srv := &http.Server{Addr: port, Handler: http.HandlerFunc(redirectHTTPport)}
	log.Info(fmt.Sprintf("Starting listen on port %s", srv.Addr))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("Port", srv.Addr, err)
		}
	}()

	stopListen(srv, shutdownCh, log)
}

// stop listen port, then come signal close application
func stopListen(srv *http.Server, shutdownCh <-chan struct{}, log *slog.Logger) {
	go func() {
		<-shutdownCh

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		log.Info(fmt.Sprintf("Shutting down server on port %s", srv.Addr))
		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Server shutdown failed on port", srv.Addr, err)
		}
	}()
}

// redirect to HTTPs
func redirectHTTPport(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.Path
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
