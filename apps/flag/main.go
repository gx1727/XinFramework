package main

import (
	"log"

	"gx1727.com/xin/apps/flag"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/repository"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	logger.Init(cfg.Log.Dir, cfg.Log.Level)

	localStorage := storage_local.NewLocalStorage("./uploads", "/uploads")
	svc := flag.NewService(localStorage)
	h := flag.NewHandler(svc)

	mod := flag.Module(h)
	log.Printf("flag app module: %s", mod.Name())
}
