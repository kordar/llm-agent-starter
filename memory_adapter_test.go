package agentstarter

import (
	"context"
	"testing"

	llmagent "github.com/kordar/llm-agent"
	llmmemory "github.com/kordar/llm-memory"
)

func TestMemoryAdapter_PersistAndBuild(t *testing.T) {
	fm := llmmemory.NewFusionMemory(
		llmmemory.NewInMemoryStore(),
		llmmemory.NewInMemoryStore(),
		nil,
		llmmemory.RuleBasedScorer{},
		llmmemory.DefaultConfig(),
	)
	adapter := NewMemoryAdapterFromFusion(fm)

	err := adapter.Persist(context.Background(), "s1", []llmagent.Message{
		{Role: "user", Content: "报销流程是什么"},
		{Role: "assistant", Content: "先申请后审批"},
	})
	if err != nil {
		t.Fatalf("Persist error: %v", err)
	}

	msgs, err := adapter.Build(context.Background(), "s1", "再说一遍")
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(msgs) == 0 {
		t.Fatal("expected non-empty messages")
	}
}

func TestMemoryAdapter_Nil(t *testing.T) {
	var adapter *MemoryAdapter
	if _, err := adapter.Build(context.Background(), "s1", "q"); err == nil {
		t.Fatal("expected build error for nil adapter")
	}
	if err := adapter.Persist(context.Background(), "s1", nil); err == nil {
		t.Fatal("expected persist error for nil adapter")
	}
}
