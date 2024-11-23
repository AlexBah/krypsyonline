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

	if port == ":80" {
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				panic(err)
			}
		}()
	}
	if port == ":443" {
		go func() {
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
				panic(err)
			}
		}()
	}

	<-shutdownCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Info(fmt.Sprintf("Shutting down server on %s", port))
	if err := srv.Shutdown(ctx); err != nil {
		log.Info(fmt.Sprintf("Server on %s shutdown failed:%v", port, err))
	}
}

// listen secondary port and redirect request to primary port
func listenSecondaryPort(port string, shutdownCh <-chan struct{}, log *slog.Logger) {
	srv := &http.Server{Addr: port, Handler: http.HandlerFunc(redirectHTTPports)}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	<-shutdownCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Info(fmt.Sprintf("Shutting down server on %s", port))
	if err := srv.Shutdown(ctx); err != nil {
		log.Info(fmt.Sprintf("Server on %s shutdown failed: %v", port, err))
	}
}

// redirect to HTTPs
func redirectHTTPports(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.Path
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
