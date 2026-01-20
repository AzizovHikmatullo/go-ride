package server

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AzizovHikmatullo/go-ride/internal/auth"
	"github.com/AzizovHikmatullo/go-ride/internal/config"
	"github.com/AzizovHikmatullo/go-ride/internal/middleware"
	"github.com/AzizovHikmatullo/go-ride/internal/rides"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type App struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sqlx.DB
	r      *gin.Engine
}

func NewApp(cfg *config.Config, db *sqlx.DB, logger *slog.Logger) *App {
	app := &App{
		cfg:    cfg,
		logger: logger,
		db:     db,
		r:      gin.Default(),
	}

	return app
}

func (a *App) InitRoutes() {
	a.r.Use(middleware.LoggerMiddleware(a.logger))

	authRepo := auth.NewRepository(a.db, a.logger)
	ridesRepo := rides.NewRepository(a.db, a.logger)

	authService := auth.NewAuthService(authRepo, a.cfg.JWT.Secret, a.cfg.JWT.AccessTokenTTL, a.cfg.JWT.RefreshTokenTTL, a.logger)
	ridesService := rides.NewRideService(ridesRepo, a.logger)

	authHandler := auth.NewAuthHandler(authService)
	ridesHandler := rides.NewRideHandler(ridesService)

	authRoutes := a.r.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.RefreshToken)
		authRoutes.POST("/logout", authHandler.Logout)
	}

	userRoutes := a.r.Group("/rides")
	userRoutes.Use(middleware.RequireRole("USER"))
	{
		userRoutes.POST("", ridesHandler.CreateRide)
		userRoutes.GET("/:id", ridesHandler.GetRideByID)
		userRoutes.GET("/:id/status", ridesHandler.GetRideStatus)
		userRoutes.POST("/:id/cancel", ridesHandler.CancelRide)

		userRoutes.POST("/search", ridesHandler.GetSearchingRides).Use(middleware.RequireRole("DRIVER"))
		userRoutes.POST("/:id/take", ridesHandler.TakeRide).Use(middleware.RequireRole("DRIVER"))
		userRoutes.POST("/:id/complete", ridesHandler.CompleteRide).Use(middleware.RequireRole("DRIVER"))
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
