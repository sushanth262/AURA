package security_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/security"
)

func TestInlineClient_RedactsEmail(t *testing.T) {
	raw := map[string]any{
		"from": "ops-alerts@example.com",
		"nested": map[string]any{
			"body": "Contact oncall@example.com for details",
		},
	}
	out, err := security.InlineClient{}.Redact(context.Background(), raw, "email")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(out)
	s := string(b)
	if security.ContainsRawEmail(s) {
		t.Fatalf("raw email remained: %s", s)
	}
	if !strings.Contains(s, "[EMAIL_REDACTED]") {
		t.Fatalf("expected redaction token: %s", s)
	}
}

func TestRedactValue_PreservesNonEmail(t *testing.T) {
	out := security.RedactValue(map[string]any{"service": "api-gateway", "count": 42})
	m := out.(map[string]any)
	if m["service"] != "api-gateway" {
		t.Fatalf("got %v", m)
	}
}
