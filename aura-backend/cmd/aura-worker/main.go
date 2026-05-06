package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/workersvc"
)

func main() {
	cfg := config.Load()
	if os.Getenv("HTTP_ADDR") == "" {
		cfg.HTTPAddr = ":8083"
	}
	srv := &workersvc.Server{Cfg: cfg}
	log.Printf("aura-worker listening on %s (WORKER_ENABLED_SOURCES=%v)",
		cfg.HTTPAddr, cfg.EnabledSources)
	if err := http.ListenAndServe(cfg.HTTPAddr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
