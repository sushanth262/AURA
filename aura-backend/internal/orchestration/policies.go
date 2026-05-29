package orchestration

// SynthesisJoinPolicy controls when retrieval may proceed to synthesis.
type SynthesisJoinPolicy string

const (
	JoinAnySuccess SynthesisJoinPolicy = "any_success"
	JoinAllRequired SynthesisJoinPolicy = "all_required"
)

// Policies holds tenant-level orchestration rules.
type Policies struct {
	MinAgentsForSynthesis int
	SynthesisJoin         SynthesisJoinPolicy
}

// DefaultPolicies matches production spec defaults for the three-agent stub path.
func DefaultPolicies() Policies {
	return Policies{
		MinAgentsForSynthesis: 1,
		SynthesisJoin:         JoinAnySuccess,
	}
}

// PoliciesFromConfig maps env strings to Policies.
func PoliciesFromConfig(minAgents int, synthesisJoin string) Policies {
	p := DefaultPolicies()
	if minAgents > 0 {
		p.MinAgentsForSynthesis = minAgents
	}
	switch synthesisJoin {
	case string(JoinAllRequired):
		p.SynthesisJoin = JoinAllRequired
	}
	return p
}

// CanSynthesize reports whether enough agent nodes succeeded to enter synthesis.
func (p Policies) CanSynthesize(successCount, enabledAgentCount int) bool {
	if successCount < p.MinAgentsForSynthesis {
		return false
	}
	switch p.SynthesisJoin {
	case JoinAllRequired:
		return successCount >= enabledAgentCount
	default:
		return successCount >= p.MinAgentsForSynthesis
	}
}
