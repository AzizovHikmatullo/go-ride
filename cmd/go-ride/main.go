package main

import (
	"log/slog"

	"github.com/AzizovHikmatullo/go-ride/internal/config"
	"github.com/AzizovHikmatullo/go-ride/internal/db"
	"github.com/AzizovHikmatullo/go-ride/internal/logger"
	"github.com/AzizovHikmatullo/go-ride/internal/server"
)

func main() {
	logger := logger.NewLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		return
	}

	db, err := db.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect to db", slog.String("error", err.Error()))
		return
	}

	app := server.NewApp(cfg, db, logger)

	app.Run()
}
