package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// NewAgentWorkerClient builds an HTTP client for POST /v1/agents/{domain}/execute.
func NewAgentWorkerClient(cfg config.Config) orchestration.AgentWorkerClient {
	base := strings.TrimSpace(cfg.WorkerURL)
	if base == "" {
		return noopAgentWorker{}
	}
	return &httpAgentWorkerClient{
		base:   strings.TrimRight(base, "/"),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type noopAgentWorker struct{}

func (noopAgentWorker) Execute(context.Context, orchestration.AgentDomain, orchestration.AgentTask) (orchestration.AgentResult, error) {
	return orchestration.AgentResult{Status: orchestration.AgentFailed}, fmt.Errorf("worker URL not configured")
}

type httpAgentWorkerClient struct {
	base   string
	client *http.Client
}

func (c *httpAgentWorkerClient) Execute(ctx context.Context, domain orchestration.AgentDomain, task orchestration.AgentTask) (orchestration.AgentResult, error) {
	domain = orchestration.AgentDomain(strings.ToLower(strings.TrimSpace(string(domain))))
	task.Domain = domain

	joined, err := url.JoinPath(c.base, "v1", "agents", string(domain), "execute")
	if err != nil {
		return orchestration.AgentResult{}, err
	}
	body, err := json.Marshal(task)
	if err != nil {
		return orchestration.AgentResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joined, bytes.NewReader(body))
	if err != nil {
		return orchestration.AgentResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return orchestration.AgentResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return orchestration.AgentResult{Status: orchestration.AgentFailed},
			fmt.Errorf("worker execute %s returned %s", domain, resp.Status)
	}
	var result orchestration.AgentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return orchestration.AgentResult{}, err
	}
	return result, nil
}
