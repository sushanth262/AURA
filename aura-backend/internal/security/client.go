package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
)

// HTTPClient calls an external Security & Redaction service.
type HTTPClient struct {
	base   string
	fallback InlineClient
	client *http.Client
}

// NewHTTPClient posts payloads to SECURITY_SERVICE_URL/v1/redact.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		base:     strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		fallback: InlineClient{},
		client:   &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *HTTPClient) Redact(ctx context.Context, raw map[string]any, source string) (map[string]any, error) {
	if raw == nil {
		return nil, nil
	}
	if c.base == "" {
		return c.fallback.Redact(ctx, raw, source)
	}
	body, _ := json.Marshal(map[string]any{
		"source": source,
		"payload": raw,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/redact", bytes.NewReader(body))
	if err != nil {
		return c.fallback.Redact(ctx, raw, source)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return c.fallback.Redact(ctx, raw, source)
	}
	defer resp.Body.Close()
	var out struct {
		Payload map[string]any `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.Payload == nil {
		return c.fallback.Redact(ctx, raw, source)
	}
	return out.Payload, nil
}

// NewClient selects inline or HTTP security client from config.
func NewClient(cfg config.Config) Client {
	mode := strings.ToLower(strings.TrimSpace(cfg.SecurityMode))
	if mode == "http" && strings.TrimSpace(cfg.SecurityServiceURL) != "" {
		return NewHTTPClient(cfg.SecurityServiceURL)
	}
	return InlineClient{}
}

// Ensure InlineClient satisfies Client.
var _ Client = InlineClient{}

// ErrRedactionFailed indicates mandatory security stage failure.
var ErrRedactionFailed = fmt.Errorf("security redaction failed")
