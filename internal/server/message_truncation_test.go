package server

import (
	"strings"
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/streaming"
)

// makeMsg builds a streaming.Message whose Content field has exactly n characters
// so that estimateMessageTokens returns a predictable token count (n/charsPerToken,
// minimum 1).
func makeMsg(role, content string) streaming.Message {
	return streaming.Message{Role: role, Content: content}
}

func makeToolResultMsg(role, content string) streaming.Message {
	return streaming.Message{
		Role: role,
		ToolResults: []streaming.ToolResult{
			{Content: content},
		},
	}
}

// ─── estimateMessageTokens ────────────────────────────────────────────────────

func TestEstimateMessageTokens_PlainText(t *testing.T) {
	m := makeMsg("user", strings.Repeat("a", 400)) // 400 chars → 100 tokens
	got := estimateMessageTokens(m)
	if got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestEstimateMessageTokens_ToolResult(t *testing.T) {
	m := makeToolResultMsg("tool", strings.Repeat("b", 800)) // 800 chars in ToolResults → 200 tokens
	got := estimateMessageTokens(m)
	if got != 200 {
		t.Errorf("expected 200, got %d", got)
	}
}

func TestEstimateMessageTokens_MinimumOne(t *testing.T) {
	m := makeMsg("assistant", "") // 0 chars → should be clamped to 1
	got := estimateMessageTokens(m)
	if got != 1 {
		t.Errorf("expected 1 (minimum), got %d", got)
	}
}

func TestEstimateMessageTokens_ToolUse(t *testing.T) {
	m := streaming.Message{
		Role: "assistant",
		ToolUses: []streaming.ToolUse{
			{
				Name:  strings.Repeat("x", 4),  // 4 chars → 1 token for name
				Input: map[string]interface{}{"k": strings.Repeat("v", 396)}, // ~400 chars of marshalled JSON
			},
		},
	}
	got := estimateMessageTokens(m)
	// name: 4 chars, input JSON roughly 400+ chars; together > 100 tokens
	if got < 100 {
		t.Errorf("expected at least 100 tokens for a large tool-use input, got %d", got)
	}
}

// ─── truncateMessagesByTokenBudget ───────────────────────────────────────────

func TestTruncate_EmptySlice(t *testing.T) {
	got := truncateMessagesByTokenBudget(nil, 10000, nopLog)
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestTruncate_SingleMessage(t *testing.T) {
	msgs := []streaming.Message{makeMsg("user", "hello")}
	got := truncateMessagesByTokenBudget(msgs, 10000, nopLog)
	if len(got) != 1 {
		t.Errorf("expected 1 message, got %d", len(got))
	}
}

func TestTruncate_NothingDroppedWhenBudgetSufficient(t *testing.T) {
	msgs := []streaming.Message{
		makeMsg("user", strings.Repeat("a", 40)),    // 10 tokens
		makeMsg("assistant", strings.Repeat("b", 40)), // 10 tokens
		makeMsg("user", strings.Repeat("c", 40)),    // 10 tokens
	}
	got := truncateMessagesByTokenBudget(msgs, 100, nopLog)
	if len(got) != 3 {
		t.Errorf("expected 3 messages (no truncation), got %d", len(got))
	}
}

func TestTruncate_KeepsFirstAndMostRecent(t *testing.T) {
	// Create 5 messages; each is 100 tokens (400 chars).  Budget is 250 tokens.
	// First message costs 100 tokens → 150 tokens available for history.
	// Messages 2-5 are 100 tokens each.
	// Backwards fill: msg5 (used=100 ≤ 150 → kept=1), msg4 (used=200 > 150 → break).
	// Expected result: [msg1, msg5] (2 messages)
	chars400 := strings.Repeat("x", 400)
	msgs := []streaming.Message{
		makeMsg("user", chars400),      // msg1 – always kept
		makeMsg("assistant", chars400), // msg2
		makeMsg("user", chars400),      // msg3
		makeMsg("assistant", chars400), // msg4
		makeMsg("user", chars400),      // msg5
	}

	var logLines []string
	got := truncateMessagesByTokenBudget(msgs, 250, func(s string) { logLines = append(logLines, s) })

	if len(got) != 2 {
		t.Errorf("expected 2 messages (first + most-recent), got %d", len(got))
	}
	if got[0].Content != chars400 {
		t.Error("first message not preserved")
	}
	if got[len(got)-1].Content != chars400 {
		t.Error("last message should be msg5 (most recent)")
	}
	if len(logLines) == 0 {
		t.Error("expected truncation to be logged")
	}
	if !strings.Contains(logLines[0], "dropped") {
		t.Errorf("log line should mention 'dropped', got: %q", logLines[0])
	}
}

func TestTruncate_FirstMessageExceedsBudget(t *testing.T) {
	// Even if the first message alone exceeds the budget, it must be returned.
	big := strings.Repeat("z", 4000) // 1000 tokens
	msgs := []streaming.Message{
		makeMsg("user", big),
		makeMsg("assistant", "short"),
	}
	got := truncateMessagesByTokenBudget(msgs, 50, nopLog) // budget < first msg
	if len(got) != 1 {
		t.Errorf("expected only first message returned, got %d", len(got))
	}
	if got[0].Content != big {
		t.Error("returned message should be the first (large) message")
	}
}

func TestTruncate_AllHistoryFits(t *testing.T) {
	// Each message is 4 chars → 1 token.  Budget is 100 tokens → all fit.
	var msgs []streaming.Message
	for i := 0; i < 10; i++ {
		msgs = append(msgs, makeMsg("user", "abcd")) // 1 token each
	}
	var logLines []string
	got := truncateMessagesByTokenBudget(msgs, 100, func(s string) { logLines = append(logLines, s) })

	if len(got) != 10 {
		t.Errorf("expected all 10 messages, got %d", len(got))
	}
	if len(logLines) != 0 {
		t.Errorf("expected no log output when nothing is dropped, got: %v", logLines)
	}
}

func TestTruncate_DropsManyOldMessages(t *testing.T) {
	// 51 messages each 1 token.  Budget = 10 tokens.
	// First msg = 1 token → 9 tokens left for history.
	// Should keep last 9 of the 50 remaining messages.
	var msgs []streaming.Message
	for i := 0; i < 51; i++ {
		msgs = append(msgs, makeMsg("user", "abcd")) // 1 token each
	}

	var logLines []string
	got := truncateMessagesByTokenBudget(msgs, 10, func(s string) { logLines = append(logLines, s) })

	// expected: 1 (first) + 9 (recent)
	if len(got) != 10 {
		t.Errorf("expected 10 messages, got %d", len(got))
	}
	// The last element should be the original last message (index 50).
	if got[len(got)-1].Content != "abcd" {
		t.Error("last message should be the most-recent original message")
	}
	if len(logLines) == 0 {
		t.Error("expected truncation log")
	}
}

// ─── contextWindowTokens ─────────────────────────────────────────────────────

func TestContextWindowTokens_UnknownModel(t *testing.T) {
	// An unrecognised model ID must fall back to constants.MaxContextTokens.
	tokens := contextWindowTokens("unknown-model-xyz")
	// Fall back value after subtracting DefaultMaxTokens.
	// In the degenerate case where budget <= 0 it returns MaxContextTokens itself.
	if tokens <= 0 {
		t.Errorf("expected positive token budget, got %d", tokens)
	}
}

func TestContextWindowTokens_KnownModel(t *testing.T) {
	// gpt-4o-mini has a 128 000-token context window in the model registry.
	tokens := contextWindowTokens("gpt-4o-mini")
	if tokens <= 0 {
		t.Errorf("expected positive token budget for gpt-4o-mini, got %d", tokens)
	}
	// Must be less than the full 128 000 window (response buffer reserved).
	if tokens >= 128_000 {
		t.Errorf("expected budget < 128000 (response buffer reserved), got %d", tokens)
	}
}
