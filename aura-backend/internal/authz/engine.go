package authz

import (
	"context"
	"slices"
)

// Request models “Can subject X perform action Y on resource Z?” for stub policies.
type Request struct {
	Subject      string
	Action       string
	ResourceType string
	ResourceID   string
	TenantID     string
	Roles        []string
}

type Response struct {
	Allowed       bool
	Reason        string
	PolicyVersion string
	DecisionID    string
}

// Engine is the pluggable authorization surface; StubEngine is replaceable with OPA/Cerbos later.
type Engine interface {
	Authorize(ctx context.Context, policyVersion string, req Request) Response
}

// StubEngine implements simple role × action matrix (see docs/BFF_AUTH_LOGIN.md).
type StubEngine struct{}

func (StubEngine) Authorize(_ context.Context, policyVersion string, req Request) Response {
	decisionID := "" // filled by caller / audit wrapper

	has := func(role string) bool {
		return slices.Contains(req.Roles, role)
	}

	allowed := false
	reason := "deny_by_default"

	switch req.Action {
	case "incident:create":
		if has("operator") || has("admin") {
			allowed = true
			reason = "role_allows_create"
		} else if has("viewer") {
			reason = "viewer_cannot_create"
		} else {
			reason = "missing_operator_or_admin_role"
		}
	case "incident:read":
		if has("viewer") || has("operator") || has("admin") {
			allowed = true
			reason = "role_allows_read"
		} else {
			reason = "missing_any_read_role"
		}
	case "incident:history":
		if has("viewer") || has("operator") || has("admin") {
			allowed = true
			reason = "role_allows_history"
		} else {
			reason = "missing_any_read_role"
		}
	default:
		reason = "unknown_action"
	}

	return Response{
		Allowed:       allowed,
		Reason:        reason,
		PolicyVersion: policyVersion,
		DecisionID:    decisionID,
	}
}
