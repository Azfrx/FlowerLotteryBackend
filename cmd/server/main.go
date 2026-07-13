package main

import (
	"flower-lottery-backend/initialize"
	"flower-lottery-backend/router"
	"fmt"
	"go.uber.org/zap"
	"os"
)

func main() {
	cfg, err := initialize.Config()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	log, err := initialize.Logger(cfg.Log)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer log.Sync()
	db, err := initialize.Database(cfg.Database)
	if err != nil {
		log.Fatal("database init failed", zap.Error(err))
	}
	engine := router.New(db, log, cfg)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := engine.Run(addr); err != nil {
		log.Fatal("server stopped", zap.Error(err))
	}
}
