package agentstarter

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	llmagent "github.com/kordar/llm-agent"
	"github.com/spf13/cast"
)

type ClientBuilder func(cfg map[string]string) (llmagent.LLM, error)

var (
	clientBuilderMu sync.RWMutex
	clientBuilders  = make(map[string]ClientBuilder)
)

func init() {
	_ = RegisterClientBuilder("ollama", buildOllamaClient)
}

func RegisterClientBuilder(clientType string, builder ClientBuilder) error {
	clientType = strings.ToLower(strings.TrimSpace(clientType))
	if clientType == "" {
		return fmt.Errorf("client type is empty")
	}
	if builder == nil {
		return fmt.Errorf("client builder is nil for type=%s", clientType)
	}
	clientBuilderMu.Lock()
	clientBuilders[clientType] = builder
	clientBuilderMu.Unlock()
	return nil
}

func BuildClientFromConfig(cfg map[string]string) (llmagent.LLM, error) {
	clientType := strings.ToLower(strings.TrimSpace(cfg["type"]))
	if clientType == "" {
		clientType = "ollama"
	}

	clientBuilderMu.RLock()
	builder, ok := clientBuilders[clientType]
	clientBuilderMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unsupported client type: %s", clientType)
	}
	return builder(cfg)
}

// AgentItemLoader is the custom callback for each namespace item.
type AgentItemLoader func(moduleName string, itemID string, item map[string]string, agent *llmagent.Agent) error

// Module is a startup module scaffold for multi-namespace agent bootstrapping.
type Module struct {
	name     string
	loadItem AgentItemLoader
}

func NewModule(name string, load AgentItemLoader) *Module {
	if name == "" {
		name = "llm-agent"
	}
	return &Module{
		name:     name,
		loadItem: load,
	}
}

func (m *Module) Name() string { return m.name }

func (m *Module) Provide(id string) (*llmagent.Agent, error) {
	return GetE(id)
}

func (m *Module) Load(value interface{}) {
	items := cast.ToStringMap(value)
	if items["id"] != nil {
		id := cast.ToString(items["id"])
		m.loadOne(id, cast.ToStringMapString(value))
		return
	}

	for key, item := range items {
		m.loadOne(key, cast.ToStringMapString(item))
	}
}

func (m *Module) Close() {
}

func (m *Module) loadOne(id string, cfg map[string]string) error {
	if id == "" {
		return fmt.Errorf("%s: attribute id cannot be empty", m.Name())
	}

	client, err := buildClientByConfig(cfg)
	if err != nil {
		return fmt.Errorf("%s: build client failed for id=%s: %w", m.Name(), id, err)
	}
	a := llmagent.NewAgent(client)
	Provide(id, a)

	if m.loadItem != nil {
		if err := m.loadItem(m.name, id, cfg, a); err != nil {
			return fmt.Errorf("%s: custom loader failed for id=%s: %w", m.Name(), id, err)
		}
		slog.Debug("triggering custom loader completion", "module", m.Name(), "id", id)
	}
	slog.Info("loading module successfully", "module", m.Name(), "id", id)
	return nil
}

func buildClientByConfig(cfg map[string]string) (llmagent.LLM, error) {
	return BuildClientFromConfig(cfg)
}

func buildOllamaClient(cfg map[string]string) (llmagent.LLM, error) {
	endpoint := strings.TrimSpace(cfg["endpoint"])
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	c := llmagent.NewOllamaClient(endpoint)

	// Support multiple header config styles for flexible startup config.
	headers := map[string]string{}
	if raw := strings.TrimSpace(cfg["headers"]); raw != "" {
		if err := json.Unmarshal([]byte(raw), &headers); err != nil {
			return nil, fmt.Errorf("invalid headers json: %w", err)
		}
	}
	slog.Info("headers parsed", "headers", headers)
	if token := strings.TrimSpace(cfg["bearer_token"]); token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	if auth := strings.TrimSpace(cfg["authorization"]); auth != "" {
		headers["Authorization"] = auth
	}
	if len(headers) > 0 {
		c.Headers = headers
	}
	return c, nil
}
