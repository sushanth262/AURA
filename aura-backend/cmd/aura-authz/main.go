package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sushanth262/AURA/aura-backend/internal/authz"
	"github.com/sushanth262/AURA/aura-backend/internal/authzsvc"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
)

func main() {
	cfg := config.Load()
	if os.Getenv("HTTP_ADDR") == "" {
		cfg.HTTPAddr = ":8081"
	}
	srv := &authzsvc.Server{
		Cfg:    cfg,
		Engine: authz.StubEngine{},
	}
	log.Printf("aura-authz listening on %s (POLICY_VERSION=%s)", cfg.HTTPAddr, cfg.PolicyVersion)
	if err := http.ListenAndServe(cfg.HTTPAddr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}