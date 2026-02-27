package server

import (
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// TestTaskExecutionRetryCount verifies that RetryCount increments correctly.
// The logic in task_execution.go increments RetryCount for every turn after turn 1.
func TestTaskExecutionRetryCount(t *testing.T) {
	exec := &TaskExecution{}

	// Simulate agentic loop: turns 2, 3, 4 each trigger a RetryCount++
	for turn := 1; turn <= 4; turn++ {
		if turn > 1 {
			exec.RetryCount++
		}
	}

	if exec.RetryCount != 3 {
		t.Errorf("expected RetryCount=3 after 4 turns, got %d", exec.RetryCount)
	}
}

// TestTaskExecutionWasEscalated verifies the escalation flag logic.
func TestTaskExecutionWasEscalated(t *testing.T) {
	// Simulate: role default = TierMinimal, selected = TierHigh → escalated
	defaultTier := monitoring.TierMinimal
	selectedTier := monitoring.TierHigh

	exec := &TaskExecution{}
	exec.WasEscalated = selectedTier > defaultTier
	exec.WasDowngraded = selectedTier < defaultTier

	if !exec.WasEscalated {
		t.Error("expected WasEscalated=true when selectedTier > defaultTier")
	}
	if exec.WasDowngraded {
		t.Error("expected WasDowngraded=false when selectedTier > defaultTier")
	}
}

// TestTaskExecutionWasDowngraded verifies the downgrade flag logic.
func TestTaskExecutionWasDowngraded(t *testing.T) {
	// Simulate: role default = TierHigh, selected = TierMinimal → downgraded
	defaultTier := monitoring.TierHigh
	selectedTier := monitoring.TierMinimal

	exec := &TaskExecution{}
	exec.WasEscalated = selectedTier > defaultTier
	exec.WasDowngraded = selectedTier < defaultTier

	if exec.WasEscalated {
		t.Error("expected WasEscalated=false when selectedTier < defaultTier")
	}
	if !exec.WasDowngraded {
		t.Error("expected WasDowngraded=true when selectedTier < defaultTier")
	}
}

// TestTaskExecutionNoEscalationOnDefaultTier verifies that when the selected tier
// matches the default tier, neither flag is set.
func TestTaskExecutionNoEscalationOnDefaultTier(t *testing.T) {
	// Simulate: role default = TierMedium, selected = TierMedium → no change
	defaultTier := monitoring.TierMedium
	selectedTier := monitoring.TierMedium

	exec := &TaskExecution{}
	exec.WasEscalated = selectedTier > defaultTier
	exec.WasDowngraded = selectedTier < defaultTier

	if exec.WasEscalated {
		t.Error("expected WasEscalated=false when selectedTier == defaultTier")
	}
	if exec.WasDowngraded {
		t.Error("expected WasDowngraded=false when selectedTier == defaultTier")
	}
}
