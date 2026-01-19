package main

import (
	"fmt"
	"log"

	"github.com/AzizovHikmatullo/go-ride/internal/server"
	"github.com/AzizovHikmatullo/go-ride/pkg/config"
	"github.com/AzizovHikmatullo/go-ride/pkg/db"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	fmt.Println(cfg.Database.Host, cfg.Database.Password, cfg.Database.Port, cfg.Database.User)

	db, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err)
	}

	app := server.NewApp(cfg, db)

	app.Run()
}
