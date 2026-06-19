package app

import (
	"context"
	"strings"
)

const queryRewriteSystemPrompt = `You reformulate a user's latest message into a short, self-contained search query for a document-retrieval system.
- Resolve pronouns and references using the prior conversation so the query stands alone.
- Preserve verbatim every name, ID, date, amount, invoice or contract number, and proper noun.
- Remove filler and conversational wording.
- Reply with ONLY the search query: no quotes, no preamble, no explanation.
- Keep the query in the same language as the user's message.
If the latest message is already a clear standalone query, echo it unchanged.`

// maxRewriteHistory caps how many prior turns the rewriter sees, keeping the
// reformulation cheap and focused on recent context.
const maxRewriteHistory = 6

// rewriteQuery produces a standalone retrieval query from the conversation.
// It is best-effort: any error or empty result falls back to rawQuery, so
// retrieval never hard-fails on a rewriting problem. Single-turn conversations
// return rawQuery unchanged — the latest message is already standalone, so the
// extra LLM round-trip is skipped.
func rewriteQuery(ctx context.Context, llm LLMChatPort, model, apiKey, rawQuery string, messages []ChatMessage) string {
	if len(messages) <= 1 {
		return rawQuery
	}

	// messages ends with the current user turn; feed everything before it as
	// context and pass the latest turn explicitly to avoid duplication.
	prior := messages[:len(messages)-1]
	if len(prior) > maxRewriteHistory {
		prior = prior[len(prior)-maxRewriteHistory:]
	}

	llmMessages := make([]LLMMessage, 0, len(prior)+2)
	llmMessages = append(llmMessages, LLMMessage{Role: "system", Content: queryRewriteSystemPrompt})
	for _, m := range prior {
		if m.Role != "user" && m.Role != "assistant" {
			continue
		}
		llmMessages = append(llmMessages, LLMMessage{Role: m.Role, Content: m.Content})
	}
	llmMessages = append(llmMessages, LLMMessage{
		Role:    "user",
		Content: "Latest message: " + rawQuery + "\n\nStandalone search query:",
	})

	out, err := complete(ctx, llm, LLMChatRequest{Model: model, APIKey: apiKey, Messages: llmMessages})
	if err != nil || strings.TrimSpace(out) == "" {
		return rawQuery
	}
	return strings.TrimSpace(out)
}

// complete returns the full (non-streamed) completion for a request by
// accumulating the provider's streamed deltas into a single string.
func complete(ctx context.Context, llm LLMChatPort, req LLMChatRequest) (string, error) {
	var b strings.Builder
	err := llm.StreamCompletion(ctx, req, nil, func(delta string) {
		b.WriteString(delta)
	})
	return strings.TrimSpace(b.String()), err
}
