package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"

	"gopkg.in/yaml.v3"
)

const (
	DefaultRepository = "any"
)

type SettingsManager struct {
	mapping        map[string]*RepositorySettings
	defaultSetting RepositorySettings
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
	if settings, ok := m.mapping[repository]; ok {
		return settings
	}
	return &m.defaultSetting
}


func SetupSettingsManager() (*SettingsManager, error) {
	settingsDir := config.GlobalConfig.RepositoriesDirPath
	entries, err := os.ReadDir(settingsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings dir %s: %w", settingsDir, err)
	}

	manager := &SettingsManager{
		mapping: make(map[string]*RepositorySettings),
	}

	var defaultSettingsInitialized bool
	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			logger.Debug(nil, "directory skipped", "action", "settings_setup", "name", name)
			continue
		}
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			logger.Debug(nil, "config file skipped", "action", "settings_setup", "name", name)
			continue
		}

		path := filepath.Join(settingsDir, name)
		settings, err := loadRepositorySettings(path)
		if err != nil {
			return nil, err
		}

		if settings.Repository == DefaultRepository {
			manager.defaultSetting = *settings
			defaultSettingsInitialized = true
			logger.Debug(nil, "loaded default repository settings",
				"action", "settings_setup",
				"name", name,
				"actions", manager.defaultSetting.RenderActionsCodes())
			continue
		}

		manager.mapping[settings.Repository] = settings
		logger.Debug(nil, "loaded repository settings",
			"action", "settings_setup",
			"name", name,
			"repository", settings.Repository,
			"actions", settings.RenderActionsCodes())
	}

	if !defaultSettingsInitialized {
		return nil, fmt.Errorf("default repository config (repository: any) is missing")
	}

	return manager, nil
}

func loadRepositorySettings(path string) (*RepositorySettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var settings RepositorySettings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse repository settings %s: %w", path, err)
	}

	if settings.Repository == "" {
		return nil, fmt.Errorf("repository name is required in %s", path)
	}

	for i, code := range settings.AllowedJiraProjects {
		settings.AllowedJiraProjects[i] = strings.ToUpper(strings.TrimSpace(code))
	}

	for i := range settings.Actions {
		eventType, ok := events.ParseEventType(settings.Actions[i].OnCode)
		if !ok {
			return nil, fmt.Errorf("action for unknown event '%s' in %q", settings.Actions[i].OnCode, path)
		}
		settings.Actions[i].On = eventType
	}

	for i := range settings.Batches {
		eventType, ok := events.ParseEventType(settings.Batches[i].EventRaw)
		if !ok {
			return nil, fmt.Errorf("batch for unknown event '%s' in %q", settings.Batches[i].EventRaw, path)
		}
		settings.Batches[i].Event = eventType
	}

	return &settings, nil
}
