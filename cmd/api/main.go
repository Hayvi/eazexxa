package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/betpro/server/internal/config"
	"github.com/betpro/server/internal/db"
	"github.com/betpro/server/internal/handlers"
	"github.com/betpro/server/internal/middleware"
	"github.com/betpro/server/internal/server"
	"github.com/betpro/server/internal/services"
	"github.com/betpro/server/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.NodeEnv)
	logger.Info("Configuration loaded successfully",
		"port", cfg.Port,
		"env", cfg.NodeEnv,
	)

	ctx := context.Background()

	database, err := db.New(ctx, cfg)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	redisClient, err := services.NewRedisClient(cfg)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	authService := services.NewAuthService(cfg)
	userService := services.NewUserService(database)
	profileCache := services.NewRedisProfileCache(redisClient)

	authHandler := handlers.NewAuthHandler(authService, userService, profileCache)
	userHandler := handlers.NewUserHandler(userService)

	router := server.NewRouter()
	router.Use(
		middleware.Recovery,
		middleware.Logging,
		middleware.CORS(cfg.CORSOrigins),
	)

	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	router.HandleFunc("POST /api/auth/register", authHandler.Register)
	router.HandleFunc("POST /api/auth/login", authHandler.Login)

	userGroup := router.Group("/api/users", middleware.Auth(authService, profileCache))
	userGroup.HandleFunc("GET /profile", userHandler.GetProfile)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("API server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server stopped")
}
