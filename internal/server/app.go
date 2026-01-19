package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AzizovHikmatullo/go-ride/internal/auth"
	"github.com/AzizovHikmatullo/go-ride/internal/middleware"
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

func (a *App) InitRoutes() {
	authRepo := auth.NewRepository(a.db, a.cfg.JWT.Secret, a.cfg.JWT.AccessTokenTTL, a.cfg.JWT.RefreshTokenTTL)

	authService := auth.NewAuthService(authRepo)

	authHandler := auth.NewAuthHandler(authService)

	authRoutes := a.r.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.RefreshToken)
		authRoutes.POST("/logout", authHandler.Logout)
	}

	apiRoutes := a.r.Group("/api")
	apiRoutes.Use(middleware.AuthMiddleware("DRIVER"))
	{
		apiRoutes.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "you have access!"})
		})
	}
}

func (a *App) Run() {
	a.InitRoutes()

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

	a.db.Close()

	log.Println("Server exiting")
}
