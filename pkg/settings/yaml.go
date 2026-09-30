package settings

import (
	"fmt"
	"os"
	"strings"

	"gitverse-notifier/pkg/events"

	"gopkg.in/yaml.v3"
)

// LoadRepositorySettingsFile reads and normalizes repository settings from a YAML file.
func LoadRepositorySettingsFile(path string) (*RepositorySettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return ParseRepositorySettings(data, path)
}

// ParseRepositorySettings unmarshals YAML bytes into normalized repository settings.
func ParseRepositorySettings(data []byte, source string) (*RepositorySettings, error) {
	var settings RepositorySettings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse repository settings %s: %w", source, err)
	}

	normalized, err := NormalizeRepositorySettings(settings)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}
	return &normalized, nil
}

// NormalizeRepositorySettings validates and normalizes fields used at runtime.
func NormalizeRepositorySettings(settings RepositorySettings) (RepositorySettings, error) {
	if settings.Repository == "" {
		return RepositorySettings{}, fmt.Errorf("repository name is required")
	}

	for i, code := range settings.AllowedJiraProjects {
		settings.AllowedJiraProjects[i] = strings.ToUpper(strings.TrimSpace(code))
	}

	if settings.TelegramTags == nil {
		settings.TelegramTags = map[string]string{}
	}

	for i := range settings.Actions {
		onCode := settings.Actions[i].OnCode
		if onCode == "" && settings.Actions[i].On != "" {
			onCode = string(settings.Actions[i].On)
			settings.Actions[i].OnCode = onCode
		}
		eventType, ok := events.ParseEventType(onCode)
		if !ok {
			return RepositorySettings{}, fmt.Errorf("action for unknown event '%s'", onCode)
		}
		settings.Actions[i].On = eventType
		settings.Actions[i].OnCode = onCode
	}

	for i := range settings.Groups {
		eventRaw := settings.Groups[i].EventRaw
		if eventRaw == "" && settings.Groups[i].Event != "" {
			eventRaw = string(settings.Groups[i].Event)
			settings.Groups[i].EventRaw = eventRaw
		}
		eventType, ok := events.ParseEventType(eventRaw)
		if !ok {
			return RepositorySettings{}, fmt.Errorf("group for unknown event '%s'", eventRaw)
		}
		settings.Groups[i].Event = eventType
		settings.Groups[i].EventRaw = eventRaw
	}

	return settings, nil
}
