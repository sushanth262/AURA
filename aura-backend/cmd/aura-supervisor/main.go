package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
	"github.com/sushanth262/AURA/aura-backend/internal/supervisor"
)

func main() {
	cfg := config.Load()
	if os.Getenv("HTTP_ADDR") == "" {
		cfg.HTTPAddr = ":8082"
	}
	store := supervisor.NewStore()
	hub := supervisor.NewWSHub(cfg)
	fetcher := supervisor.NewSnapshotFetcher(cfg)
	api := &supervisor.HTTPServer{
		Cfg:     cfg,
		Store:   store,
		Hub:     hub,
		Fetcher: fetcher,
	}
	log.Printf("aura-supervisor listening on %s (GRAPH_ENGINE_MODE=%q AGENT_EXECUTION_MODE=%q ENABLED_AGENTS=%v WORKER_URL=%q)",
		cfg.HTTPAddr, cfg.GraphEngineMode, cfg.AgentExecutionMode, cfg.EnabledAgents, cfg.WorkerURL)
	if err := registry.ValidateEnabledDomains(cfg.EnabledAgents); err != nil {
		log.Printf("aura-supervisor warning: ENABLED_AGENTS: %v", err)
	}
	if err := http.ListenAndServe(cfg.HTTPAddr, api.Router()); err != nil {
		log.Fatal(err)
	}
}
