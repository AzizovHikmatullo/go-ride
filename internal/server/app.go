package server

import (
	"context"
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

	_ "github.com/AzizovHikmatullo/go-ride/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
		r:      gin.New(),
	}

	return app
}

// @title Go-Ride API
// @version 1.0
// @description API Server for Go-Ride

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey UserAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey DriverAuth
// @in header
// @name Authorization
func (a *App) Run() {
	a.InitRoutes()

	srv := &http.Server{
		Addr:    ":" + a.cfg.Server.Port,
		Handler: a.r,
	}

	go func() {
		a.logger.Info("Running server", slog.String("port", a.cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Info("Failed to run server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		a.logger.Info("Server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	a.db.Close()

	a.logger.Info("Server exiting. Goodbye!")
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

	ridesGroup := a.r.Group("/rides")
	ridesGroup.Use(middleware.AuthMiddleware())
	{
		ridesGroup.GET("/search", middleware.RequireRole("DRIVER"), ridesHandler.GetSearchingRides)
		ridesGroup.POST("/:id/take", middleware.RequireRole("DRIVER"), ridesHandler.TakeRide)
		ridesGroup.POST("/:id/complete", middleware.RequireRole("DRIVER"), ridesHandler.CompleteRide)

		ridesGroup.POST("", middleware.RequireRole("USER"), ridesHandler.CreateRide)
		ridesGroup.GET("/:id", middleware.RequireRole("USER"), ridesHandler.GetRideByID)
		ridesGroup.GET("/:id/status", middleware.RequireRole("USER"), ridesHandler.GetRideStatus)
		ridesGroup.POST("/:id/cancel", middleware.RequireRole("USER"), ridesHandler.CancelRide)
	}

	a.r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	a.logger.Info("All routes created")
}
