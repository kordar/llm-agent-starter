package agentstarter

import (
	"context"
	"errors"

	llmagent "github.com/kordar/llm-agent"
	llmmemory "github.com/kordar/llm-memory"
)

// LLMAgentAdapter adapts llm-memory AgentBridge to llm-agent's AgentMemory interface.
type LLMAgentAdapter struct {
	bridge *llmmemory.AgentBridge
}

func NewLLMAgentAdapter(bridge *llmmemory.AgentBridge) *LLMAgentAdapter {
	return &LLMAgentAdapter{bridge: bridge}
}

func NewLLMAgentAdapterFromFusion(mem *llmmemory.FusionMemory) *LLMAgentAdapter {
	return NewLLMAgentAdapter(llmmemory.NewAgentBridge(mem))
}

func (a *LLMAgentAdapter) Build(ctx context.Context, sessionID string, userInput string) ([]llmagent.Message, error) {
	if a == nil || a.bridge == nil {
		return nil, errors.New("memory: nil llm-agent adapter")
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

func (a *LLMAgentAdapter) Persist(ctx context.Context, sessionID string, msgs []llmagent.Message) error {
	if a == nil || a.bridge == nil {
		return errors.New("memory: nil llm-agent adapter")
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
