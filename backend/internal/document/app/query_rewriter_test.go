package app

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// fakeRewriterLLM is a stub LLMChatPort that returns a canned completion (or an
// error), recording whether it was invoked.
type fakeRewriterLLM struct {
	reply string
	err   error
	calls int
}

func (f *fakeRewriterLLM) StreamCompletion(_ context.Context, req LLMChatRequest, _ func(), onDelta func(string)) error {
	if f.err != nil {
		return f.err
	}
	f.calls++
	// Echo the canned reply as a single delta, like a real provider would.
	onDelta(f.reply)
	return nil
}

func TestRewriteQuery_SingleTurnIsUnchanged(t *testing.T) {
	llm := &fakeRewriterLLM{reply: "should not be used"}
	got := rewriteQuery(context.Background(), llm, "m", "k", "teepee docs",
		[]ChatMessage{{Role: "user", Content: "teepee docs"}})

	if got != "teepee docs" {
		t.Fatalf("single-turn rewrite = %q, want raw query unchanged", got)
	}
	if llm.calls != 0 {
		t.Fatalf("rewriter made %d LLM call(s) on a single-turn chat; want 0", llm.calls)
	}
}

func TestRewriteQuery_MultiTurnUsesLLMOutput(t *testing.T) {
	llm := &fakeRewriterLLM{reply: "teepee invoice expiry date"}
	msgs := []ChatMessage{
		{Role: "user", Content: "show me the teepee invoice"},
		{Role: "assistant", Content: "Here is the invoice."},
		{Role: "user", Content: "what is its expiry?"},
	}
	got := rewriteQuery(context.Background(), llm, "m", "k", "what is its expiry?", msgs)

	if got != "teepee invoice expiry date" {
		t.Fatalf("multi-turn rewrite = %q, want the LLM output", got)
	}
	if llm.calls != 1 {
		t.Fatalf("rewriter made %d LLM call(s); want 1", llm.calls)
	}
}

func TestRewriteQuery_ErrorFallsBackToRaw(t *testing.T) {
	llm := &fakeRewriterLLM{err: errors.New("provider down")}
	msgs := []ChatMessage{
		{Role: "user", Content: "previous question"},
		{Role: "user", Content: "follow up"},
	}
	got := rewriteQuery(context.Background(), llm, "m", "k", "follow up", msgs)

	if got != "follow up" {
		t.Fatalf("error fallback = %q, want raw query", got)
	}
}

func TestRewriteQuery_EmptyOutputFallsBackToRaw(t *testing.T) {
	llm := &fakeRewriterLLM{reply: "   "}
	msgs := []ChatMessage{
		{Role: "user", Content: "previous question"},
		{Role: "user", Content: "follow up"},
	}
	got := rewriteQuery(context.Background(), llm, "m", "k", "follow up", msgs)

	if got != "follow up" {
		t.Fatalf("empty-output fallback = %q, want raw query", got)
	}
}

func TestComplete_AccumulatesDeltas(t *testing.T) {
	llm := &streamingLLM{deltas: []string{"Hello ", "world", "!"}}
	out, err := complete(context.Background(), llm, LLMChatRequest{})
	if err != nil {
		t.Fatalf("complete() error = %v", err)
	}
	if out != "Hello world!" {
		t.Fatalf("complete() = %q, want %q", out, "Hello world!")
	}
}

// streamingLLM emits a fixed sequence of deltas.
type streamingLLM struct{ deltas []string }

func (s *streamingLLM) StreamCompletion(_ context.Context, _ LLMChatRequest, _ func(), onDelta func(string)) error {
	for _, d := range s.deltas {
		onDelta(d)
	}
	return nil
}

var _ LLMChatPort = (*streamingLLM)(nil)
var _ LLMChatPort = (*fakeRewriterLLM)(nil)

// guard against accidental whitespace drift in the system prompt's critical
// "reply with ONLY the search query" instruction.
func TestQueryRewriteSystemPrompt_NoPreamble(t *testing.T) {
	if !strings.Contains(queryRewriteSystemPrompt, "Reply with ONLY the search query") {
		t.Fatal("rewrite system prompt must instruct a bare-query reply")
	}
}
