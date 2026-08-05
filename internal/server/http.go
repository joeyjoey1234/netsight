package server

import (
	"context"
	"fmt"
	"net/http"
	"netsight/internal/model"
	"time"
)

func startHTTP(ctx context.Context, config *model.ServerConfig, onStatus func(*model.ServerState)) error {
	port := config.Port
	if port == 0 {
		port = 8080
	}

	rootDir := config.RootDir
	if rootDir == "" {
		rootDir = "."
	}

	addr := fmt.Sprintf("%s:%d", config.Interface, port)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(rootDir)))

	// Add a status endpoint
	mux.HandleFunc("/__status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"NetSight HTTP Server","status":"running"}`))
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	onStatus(&model.ServerState{
		Type:   "http",
		Port:   port,
		Status: "running",
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		return err
	}
}
