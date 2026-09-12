package events

import (
	"fmt"
	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/logger"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultRepository = "any"
)

// Тип ActionSettings хранит описанный пользователем список действий для конкретного репозитория
type ActionSettings struct {
	Repository string       `yaml:"repository"`
	Actions    []ActionRule `yaml:"actions"`
}

// ActionsByEvent возвращает список действий, которые должны выполняться для указанного события
func (s *ActionSettings) ActionsByEvent(event Event) []ActionRule {
	actions := make([]ActionRule, 0)
	for _, rule := range s.Actions {
		matched := true

		if rule.On == event.Type {
			matched = matched && true
		}
		if rule.Branch == "" || event.Branch == rule.Branch {
			matched = matched && true
		}

		if matched {
			actions = append(actions, rule)
		}
	}

	return actions
}

func (s *ActionSettings) RenderActionsCodes() string {

	codes := make([]string, len(s.Actions))
	for _, action := range s.Actions {
		codes = append(codes, action.Action)
	}

	return strings.Join(codes, ", ")
}

type ActionsManager struct {
	mapping        map[string]*ActionSettings
	defaultSetting ActionSettings
}

func (m *ActionsManager) ActionsFor(repository string, event Event) []ActionRule {
	settings := m.ActionSettingsByRepository(repository)
	if settings == nil || event.Type == Unknown {
		return nil
	}

	matched := settings.ActionsByEvent(event)
	return matched
}

func (m *ActionsManager) ActionSettingsByRepository(repository string) *ActionSettings {
	if settings, ok := m.mapping[repository]; ok {
		return settings
	}
	return &m.defaultSetting
}

func SetupActionsManager() (*ActionsManager, error) {
	actionsDir := config.GlobalConfig.ActionDirPath
	entries, err := os.ReadDir(actionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read actions dir %s: %w", actionsDir, err)
	}

	manager := &ActionsManager{
		mapping: make(map[string]*ActionSettings),
	}

	var defaultSettingsInitialized bool
	for _, entry := range entries {
		name := entry.Name()
		
		if entry.IsDir() {
			logger.Debugf("[ActionsSetup] directory %s skipped", name)
			continue
		}
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			logger.Debugf("[ActionsSetup] file %s skipped", name)
			continue
		}

		path := filepath.Join(actionsDir, name)
		settings, err := loadActionSettings(path)
		if err != nil {
			return nil, err
		}

		if settings.Repository == DefaultRepository {
			manager.defaultSetting = *settings
			defaultSettingsInitialized = true
			continue
		}

		manager.mapping[settings.Repository] = settings
		logger.Debugf("[ActionsSetup] loaded actions settings (%s) for repository %s: %s", name, settings.Repository, settings.RenderActionsCodes())
	}

	if !defaultSettingsInitialized {
		return nil, fmt.Errorf("default actions config (repository: any) is missing")
	}

	return manager, nil
}

func loadActionSettings(path string) (*ActionSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var settings ActionSettings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse actions settings %s: %w", path, err)
	}

	for _, action := range settings.Actions {
		eventType, ok := ParseEventType(action.OnCode)

		if !ok {
			return nil, fmt.Errorf("action for unknown event '%s' in %q", action.OnCode, path)
		}

		action.On = eventType
	}

	return &settings, nil
}
