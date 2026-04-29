package agentstarter

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	llmagent "github.com/kordar/llm-agent"
)

var (
	providers = make(map[string]*llmagent.Agent)
	mu        sync.RWMutex
)

func Provide(id string, a *llmagent.Agent) {
	id = strings.TrimSpace(id)
	if id == "" {
		slog.Error("llm provider id is empty")
		panic(fmt.Errorf("llm provider id is empty"))
	}
	if a == nil {
		slog.Error("llm agent is nil", "id", id)
		panic(fmt.Errorf("llm agent %s is nil", id))
	}
	mu.Lock()
	defer mu.Unlock()
	providers[id] = a
}

func Get(id string) *llmagent.Agent {
	id = strings.TrimSpace(id)
	if id == "" {
		slog.Error("llm provider id is empty")
		panic(fmt.Errorf("llm provider id is empty"))
	}
	mu.RLock()
	defer mu.RUnlock()
	agent, ok := providers[id]
	if !ok {
		slog.Error("llm provider not exist", "id", id)
		panic(fmt.Errorf("llm provider %s not exist", id))
	}
	return agent
}

func GetE(id string) (*llmagent.Agent, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("llm provider id is empty")
	}
	mu.RLock()
	defer mu.RUnlock()
	agent, ok := providers[id]
	if !ok {
		return nil, fmt.Errorf("llm provider %s not exist", id)
	}
	return agent, nil
}

func ProvideFromConfig(id string, cfg map[string]string) (*llmagent.Agent, error) {
	client, err := buildClientByConfig(cfg)
	if err != nil {
		return nil, err
	}
	a := llmagent.NewAgent(client)
	Provide(id, a)
	return a, nil
}

func ProvideEFromConfig(id string, cfg map[string]string) *llmagent.Agent {
	a, err := ProvideFromConfig(id, cfg)
	if err != nil {
		slog.Error("provide llm provider failed", "id", id, "err", err)
		panic(fmt.Errorf("provide %s failed: %w", id, err))
	}
	return a
}
