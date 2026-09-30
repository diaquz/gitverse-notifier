package settings

import (
	"context"
	"fmt"
	"sync"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
)

const (
	DefaultRepository = "any"
)

// RepositorySettingsLoader loads all repository settings (typically from a store).
type RepositorySettingsLoader interface {
	List(ctx context.Context) ([]*RepositorySettings, error)
}

type SettingsManager struct {
	mu             sync.RWMutex
	mapping        map[string]*RepositorySettings
	defaultSetting *RepositorySettings
}

func (m *SettingsManager) IsJiraCodeAllowed(repository, code string) bool {
	settings := m.SettingsByRepository(repository)
	return settings.IsJiraCodeAllowed(code)
}

func (m *SettingsManager) ActionsFor(repository string, event *events.Event) []ActionRule {
	settings := m.SettingsByRepository(repository)
	if settings == nil || event.Type == events.Unknown {
		return nil
	}

	matched := settings.ActionsByEvent(event)
	return matched
}

func (m *SettingsManager) HasPotentialActions(repository string, event *events.Event) bool {
	settings := m.SettingsByRepository(repository)
	return settings.HasPotentialActions(event)
}

func (m *SettingsManager) SettingsByRepository(repository string) *RepositorySettings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if settings, ok := m.mapping[repository]; ok {
		return settings
	}
	return m.defaultSetting
}

// NewSettingsManagerFromLoader builds a manager from the full settings list.
func NewSettingsManagerFromLoader(ctx context.Context, loader RepositorySettingsLoader) (*SettingsManager, error) {
	manager := &SettingsManager{
		mapping: make(map[string]*RepositorySettings),
	}
	if err := manager.Reload(ctx, loader); err != nil {
		return nil, err
	}
	return manager, nil
}

// Reload replaces in-memory settings from the loader.
func (m *SettingsManager) Reload(ctx context.Context, loader RepositorySettingsLoader) error {
	repos, err := loader.List(ctx)
	if err != nil {
		return err
	}
	return m.ReplaceAll(repos)
}

// ReplaceAll atomically replaces the in-memory repository mapping.
func (m *SettingsManager) ReplaceAll(repos []*RepositorySettings) error {
	mapping := make(map[string]*RepositorySettings, len(repos))
	var defaultSetting *RepositorySettings

	for _, repo := range repos {
		if repo == nil {
			continue
		}
		normalized, err := NormalizeRepositorySettings(*repo)
		if err != nil {
			return err
		}
		copy := normalized
		if copy.Repository == DefaultRepository {
			defaultSetting = &copy
			continue
		}
		mapping[copy.Repository] = &copy
	}

	if defaultSetting == nil {
		return fmt.Errorf("default repository config (repository: %s) is missing", DefaultRepository)
	}

	m.mu.Lock()
	m.mapping = mapping
	m.defaultSetting = defaultSetting
	m.mu.Unlock()

	logger.Debug(nil, "settings manager reloaded",
		"action", "settings.reload",
		"repositories", len(mapping)+1,
		"default_actions", defaultSetting.RenderActionsCodes())

	return nil
}
