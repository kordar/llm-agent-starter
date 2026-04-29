package agentstarter

import (
	"context"
	"errors"

	llmagent "github.com/kordar/llm-agent"
	llmmemory "github.com/kordar/llm-memory"
)

// MemoryAdapter adapts llm-memory bridge to llm-agent memory contract.
type MemoryAdapter struct {
	bridge *llmmemory.AgentBridge
}

func NewMemoryAdapter(bridge *llmmemory.AgentBridge) *MemoryAdapter {
	return &MemoryAdapter{bridge: bridge}
}

func NewMemoryAdapterFromFusion(mem *llmmemory.FusionMemory) *MemoryAdapter {
	return NewMemoryAdapter(llmmemory.NewAgentBridge(mem))
}

func (a *MemoryAdapter) Build(ctx context.Context, sessionID string, userInput string) ([]llmagent.Message, error) {
	if a == nil || a.bridge == nil {
		return nil, errors.New("agent-starter: nil memory adapter")
	}
	msgs, err := a.bridge.Build(ctx, sessionID, userInput)
	if err != nil {
		return nil, err
	}
	out := make([]llmagent.Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, llmagent.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return out, nil
}

func (a *MemoryAdapter) Persist(ctx context.Context, sessionID string, msgs []llmagent.Message) error {
	if a == nil || a.bridge == nil {
		return errors.New("agent-starter: nil memory adapter")
	}
	in := make([]llmmemory.AgentMessage, 0, len(msgs))
	for _, m := range msgs {
		in = append(in, llmmemory.AgentMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return a.bridge.Persist(ctx, sessionID, in)
}
