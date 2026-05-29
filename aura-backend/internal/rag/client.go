package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// HTTPClient calls RAG_SERVICE_URL/v1/retrieve.
type HTTPClient struct {
	base     string
	fallback StubClient
	client   *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		base:     strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		fallback: StubClient{},
		client:   &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *HTTPClient) Retrieve(ctx context.Context, query orchestration.RAGQuery) ([]map[string]any, error) {
	if c.base == "" {
		return c.fallback.Retrieve(ctx, query)
	}
	body, _ := json.Marshal(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/retrieve", bytes.NewReader(body))
	if err != nil {
		return c.fallback.Retrieve(ctx, query)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return c.fallback.Retrieve(ctx, query)
	}
	defer resp.Body.Close()
	var out struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return c.fallback.Retrieve(ctx, query)
	}
	return out.Results, nil
}

// Client retrieves embedding-backed context.
type Client interface {
	Retrieve(ctx context.Context, query orchestration.RAGQuery) ([]map[string]any, error)
}

// NewClient selects stub or HTTP RAG client from config.
func NewClient(cfg config.Config) Client {
	mode := strings.ToLower(strings.TrimSpace(cfg.RAGMode))
	if mode == "http" && strings.TrimSpace(cfg.RAGServiceURL) != "" {
		return NewHTTPClient(cfg.RAGServiceURL)
	}
	return StubClient{}
}

var _ Client = StubClient{}
