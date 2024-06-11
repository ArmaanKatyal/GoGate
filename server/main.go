package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	opts := PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := NewPrettyHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	LoadConf()
	rh := NewRequestHandler()
	router := InitializeRoutes(rh)
	server := &http.Server{
		Addr:         ":" + AppConfig.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(AppConfig.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(AppConfig.Server.WriteTimeout) * time.Second,
	}
	slog.Info("API Gateway started", "port", AppConfig.Server.Port)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			slog.Error("Error starting server", "error", err.Error())
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(AppConfig.Server.GracefulTimeout)*time.Second)
	defer cancel()
	slog.Info("Gracefully shutting down server")
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error shutting down server", "error", err.Error())
		os.Exit(1)
	}
}
