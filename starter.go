package agentstarter

import (
	"fmt"
	"log/slog"
	"sync"
)

// AgentItemLoader is the custom callback for each namespace item.
type AgentItemLoader func(moduleName string, itemID string, item map[string]interface{}) error

// Module is a startup module scaffold for multi-namespace agent bootstrapping.
type Module struct {
	mu               sync.Mutex
	name             string
	loadItem         AgentItemLoader
	defaultNamespace string
	loadedNamespaces []string
	closed           bool
}

func NewModule(name string, load AgentItemLoader) *Module {
	if name == "" {
		name = "agent"
	}
	return &Module{
		name:             name,
		loadItem:         load,
		defaultNamespace: "default",
	}
}

func (m *Module) Name() string { return m.name }

func (m *Module) Depends() []string { return []string{} }

func (m *Module) BeforeLoad() error { return nil }

func (m *Module) Load(value interface{}) error {
	if value == nil {
		m.mu.Lock()
		ns := m.name
		if ns == "" {
			ns = "agent"
		}
		m.defaultNamespace = ns
		m.mu.Unlock()
		return m.loadOne(ns, map[string]interface{}{})
	}

	items, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%s: invalid config type %T", m.Name(), value)
	}
	if idRaw, found := items["id"]; found {
		id, ok := idRaw.(string)
		if !ok || id == "" {
			return fmt.Errorf("%s: id must be non-empty string", m.Name())
		}
		m.mu.Lock()
		m.defaultNamespace = id
		m.mu.Unlock()
		return m.loadOne(id, items)
	}

	for key, item := range items {
		cfg, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s: namespace %q config must be map[string]interface{}", m.Name(), key)
		}
		if err := m.loadOne(key, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) AfterLoad() error { return nil }

func (m *Module) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	return nil
}

func (m *Module) DefaultNamespace() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.defaultNamespace
}

func (m *Module) LoadedNamespaces() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.loadedNamespaces))
	copy(out, m.loadedNamespaces)
	return out
}

func (m *Module) loadOne(id string, cfg map[string]interface{}) error {
	if id == "" {
		return fmt.Errorf("%s: attribute id cannot be empty", m.Name())
	}
	if isDefault, _ := cfg["id_default"].(bool); isDefault {
		m.mu.Lock()
		m.defaultNamespace = id
		m.mu.Unlock()
	}
	if m.loadItem != nil {
		if err := m.loadItem(m.name, id, cfg); err != nil {
			return fmt.Errorf("%s: custom loader failed for id=%s: %w", m.Name(), id, err)
		}
		slog.Debug("triggering custom loader completion", "module", m.Name(), "id", id)
	}
	m.mu.Lock()
	m.loadedNamespaces = append(m.loadedNamespaces, id)
	m.mu.Unlock()
	slog.Info("loading module successfully", "module", m.Name(), "id", id)
	return nil
}
