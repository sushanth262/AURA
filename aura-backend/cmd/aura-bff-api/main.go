package main

import (
	"log"
	"net/http"

	"github.com/sushanth262/AURA/aura-backend/internal/bff"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
)

func main() {
	cfg := config.Load()
	srv := bff.NewServer(cfg)
	log.Printf("aura-bff-api listening on %s (AUTH_DEV_MOCK=%v AUTHZ_URL=%s SUPERVISOR_URL=%s)",
		cfg.HTTPAddr, cfg.AuthDevMock, cfg.AuthzURL, cfg.SupervisorURL)
	if err := http.ListenAndServe(cfg.HTTPAddr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
