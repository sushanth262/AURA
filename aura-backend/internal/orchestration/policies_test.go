package orchestration

import "testing"

func TestPolicies_CanSynthesize(t *testing.T) {
	p := DefaultPolicies()
	if !p.CanSynthesize(1, 3) {
		t.Fatal("expected any_success with 1 of 3")
	}
	if p.CanSynthesize(0, 3) {
		t.Fatal("expected false with 0 successes")
	}

	all := Policies{MinAgentsForSynthesis: 1, SynthesisJoin: JoinAllRequired}
	if all.CanSynthesize(2, 3) {
		t.Fatal("all_required should need 3 of 3")
	}
	if !all.CanSynthesize(3, 3) {
		t.Fatal("all_required should pass with 3 of 3")
	}
}

func TestPoliciesFromConfig(t *testing.T) {
	p := PoliciesFromConfig(2, "all_required")
	if p.MinAgentsForSynthesis != 2 {
		t.Fatalf("min agents: got %d", p.MinAgentsForSynthesis)
	}
	if p.SynthesisJoin != JoinAllRequired {
		t.Fatalf("join: got %q", p.SynthesisJoin)
	}
}
