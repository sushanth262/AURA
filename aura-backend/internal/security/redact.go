package security

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
)

var emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

const emailReplacement = "[EMAIL_REDACTED]"

// Client redacts connector payloads before findings or LLM processing.
type Client interface {
	Redact(ctx context.Context, raw map[string]any, source string) (map[string]any, error)
}

// InlineClient applies built-in redaction rules (PII-001 email).
type InlineClient struct{}

func (InlineClient) Redact(_ context.Context, raw map[string]any, _ string) (map[string]any, error) {
	if raw == nil {
		return nil, nil
	}
	out := RedactValue(raw)
	m, ok := out.(map[string]any)
	if !ok {
		return raw, nil
	}
	return m, nil
}

// RedactValue walks arbitrary JSON-like values and masks email addresses.
func RedactValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = RedactValue(val)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = RedactValue(val)
		}
		return out
	case string:
		return emailPattern.ReplaceAllString(x, emailReplacement)
	default:
		return v
	}
}

// ContainsRawEmail reports whether s contains an unredacted email pattern.
func ContainsRawEmail(s string) bool {
	return emailPattern.MatchString(s) && !strings.Contains(s, emailReplacement)
}

// CloneMap deep-copies a map via JSON for safe redaction input.
func CloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	b, err := json.Marshal(in)
	if err != nil {
		return in
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return in
	}
	return out
}
