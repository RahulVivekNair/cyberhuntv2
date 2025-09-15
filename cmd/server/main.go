package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
)

type application struct {
	logger *slog.Logger
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	app := &application{
		logger: logger,
	}
	router := app.NewRouter()

	srv := &http.Server{
		Addr:    *addr,
		Handler: router,
	}

	logger.Info("Starting server", "address", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Failed to start server", "error", err)
	}
}
