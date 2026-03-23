package main

import (
	"fmt"
	"os"

	"github.com/betpro/server/internal/config"
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
		"db_host", cfg.DBHost,
		"db_name", cfg.DBName,
	)

	logger.Info("API server starting", "port", cfg.Port)
	// TODO: Start HTTP server
}
