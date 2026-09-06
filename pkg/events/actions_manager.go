package events

import (
	"fmt"
	"gitverse-notifier/pkg/config"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ActionSettings struct {
	Repository string
	Actions    []ActionRule
}

func (s *ActionSettings) ActionsByEvent(event EventType) []ActionRule {
	if s == nil {
		return nil
	}
	out := make([]ActionRule, 0)
	for _, rule := range s.Actions {
		if rule.On == event {
			out = append(out, rule)
		}
	}
	return out
}

type ActionsManager struct {
	mapping        map[string]*ActionSettings
	defaultSetting ActionSettings
}

type actionsFileYAML struct {
	Repository string          `yaml:"repository"`
	Actions    []actionRuleYAML `yaml:"actions"`
}

type actionRuleYAML struct {
	On        string        `yaml:"on"`
	Action    string        `yaml:"action"`
	Branch    stringOrSlice `yaml:"branch"`
	Template  string        `yaml:"template"`
	SkipEmpty bool          `yaml:"skip_empty"`
}

type stringOrSlice []string

func (s *stringOrSlice) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var single string
		if err := value.Decode(&single); err != nil {
			return err
		}
		if single != "" {
			*s = []string{single}
		}
		return nil
	case yaml.SequenceNode:
		var many []string
		if err := value.Decode(&many); err != nil {
			return err
		}
		*s = many
		return nil
	case yaml.AliasNode:
		return s.UnmarshalYAML(value.Alias)
	default:
		return fmt.Errorf("branch must be a string or list of strings")
	}
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
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}

		path := filepath.Join(actionsDir, name)
		settings, err := loadActionSettings(path)
		if err != nil {
			return nil, err
		}

		repoKey := normalizeRepoKey(settings.Repository)
		if repoKey == "any" {
			manager.defaultSetting = *settings
			defaultSettingsInitialized = true
			continue
		}
		manager.mapping[repoKey] = settings
	}

	if !defaultSettingsInitialized {
		return nil, fmt.Errorf("default actions config (repository: any) is required")
	}

	return manager, nil
}

func loadActionSettings(path string) (*ActionSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var file actionsFileYAML
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("dailed to parse %s: %w", path, err)
	}

	settings := &ActionSettings{
		Repository: file.Repository,
		Actions:    make([]ActionRule, 0, len(file.Actions)),
	}

	for _, raw := range file.Actions {
		eventType, ok := ParseEventType(raw.On)
		if !ok {
			return nil, fmt.Errorf("actions for unknown event '%s' in %q", raw.On, path)
		}
		settings.Actions = append(settings.Actions, ActionRule{
			On:        eventType,
			Action:    strings.TrimSpace(raw.Action),
			Branches:  []string(raw.Branch),
			Template:  raw.Template,
			SkipEmpty: raw.SkipEmpty,
		})
	}

	return settings, nil
}

func (m *ActionsManager) ActionSettingsByRepository(repository string) *ActionSettings {
	if settings, ok := m.mapping[normalizeRepoKey(repository)]; ok {
		return settings
	}
	return &m.defaultSetting
}

func (m *ActionsManager) ActionsFor(repository string, ev Event) []ActionRule {
	settings := m.ActionSettingsByRepository(repository)
	if settings == nil || ev.Type == Unknown {
		return nil
	}

	matched := settings.ActionsByEvent(ev.Type)
	if len(matched) == 0 {
		return nil
	}

	branch := BranchName(ev.Ref)
	out := make([]ActionRule, 0, len(matched))
	for _, rule := range matched {
		if len(rule.Branches) == 0 || branchMatches(branch, rule.Branches) {
			out = append(out, rule)
		}
	}
	return out
}

func branchMatches(branch string, allowed []string) bool {
	for _, want := range allowed {
		if branch == strings.TrimSpace(want) {
			return true
		}
	}
	return false
}

func normalizeRepoKey(repo string) string {
	return strings.ToLower(strings.TrimSpace(repo))
}
