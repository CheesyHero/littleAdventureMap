package simulation

import "testing"

func TestCurrentTargetUsesCurrentDestination(t *testing.T) {
	agent := &AgentState{
		AgentID: 1,
		Destinations: [][2]float64{
			{0, 0},
			{10, 0},
			{10, 10},
		},
		PathIndex: 1,
	}

	target, err := currentTarget(agent)
	if err != nil {
		t.Fatalf("currentTarget returned error: %v", err)
	}
	if target != agent.Destinations[1] {
		t.Fatalf("expected outbound target %v, got %v", agent.Destinations[1], target)
	}
}

func TestCurrentTargetUsesHomeWhenReturningHomeEvenWithValidIndex(t *testing.T) {
	agent := &AgentState{
		AgentID:       2,
		HomePosition:  [2]float64{5, 5},
		Destinations:  [][2]float64{{0, 0}, {10, 0}},
		PathIndex:     1,
		ReturningHome: true,
	}

	target, err := currentTarget(agent)
	if err != nil {
		t.Fatalf("currentTarget returned error: %v", err)
	}
	if target != agent.HomePosition {
		t.Fatalf("expected home position %v, got %v", agent.HomePosition, target)
	}
}

func TestAdvanceRouteBeginsReverseTripAfterFinalOutbound(t *testing.T) {
	agent := &AgentState{
		Destinations: [][2]float64{
			{0, 0},
			{10, 0},
			{10, 10},
		},
		PathIndex: 2,
	}

	advanceRoute(agent)

	if !agent.ReturningHome {
		t.Fatalf("expected agent to begin returning home")
	}
	if agent.PathIndex != 1 {
		t.Fatalf("expected path index 1 after final outbound reach, got %d", agent.PathIndex)
	}
}

func TestCurrentTargetUsesHomeWhenReturningWithSentinelIndex(t *testing.T) {
	agent := &AgentState{
		AgentID:       2,
		HomePosition:  [2]float64{5, 5},
		Destinations:  [][2]float64{{0, 0}, {10, 0}},
		PathIndex:     len([][2]float64{{0, 0}, {10, 0}}),
		ReturningHome: true,
	}

	target, err := currentTarget(agent)
	if err != nil {
		t.Fatalf("currentTarget returned error: %v", err)
	}
	if target != agent.HomePosition {
		t.Fatalf("expected home position %v, got %v", agent.HomePosition, target)
	}
}
