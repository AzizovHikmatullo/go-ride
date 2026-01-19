package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AzizovHikmatullo/go-ride/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type App struct {
	cfg *config.Config
	db  *sqlx.DB
	r   *gin.Engine
}

func NewApp(cfg *config.Config, db *sqlx.DB) *App {
	app := &App{
		cfg: cfg,
		db:  db,
		r:   gin.Default(),
	}

	return app
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:    ":" + a.cfg.Server.Port,
		Handler: a.r,
	}

	go func() {
		log.Printf("Server running on port %s", a.cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
