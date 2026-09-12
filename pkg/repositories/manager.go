package repositories

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/logger"

	"gopkg.in/yaml.v3"
)

const (
	DefaultRepository = "any"
)

// RepositorySettings хранит настройки конкретного репозитория
type RepositorySettings struct {
	Repository          string   `yaml:"repository"`
	Url                 string   `yaml:"url"`
	AllowedJiraProjects []string `yaml:"allowed_jira_projects"`
}

// IsJiraCodeAllowed проверяет, разрешён ли код проекта Jira
func (s *RepositorySettings) IsJiraCodeAllowed(code string) bool {
	if len(s.AllowedJiraProjects) == 0 {
		return true
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	for _, allowed := range s.AllowedJiraProjects {
		if strings.HasPrefix(code, allowed) {
			return true
		}
	}

	return false
}

type RepositoriesManager struct {
	mapping        map[string]*RepositorySettings
	defaultSetting RepositorySettings
}

func (m *RepositoriesManager) IsJiraCodeAllowed(repository, code string) bool {
	settings := m.SettingsByRepository(repository)
	return settings.IsJiraCodeAllowed(code)
}

func (m *RepositoriesManager) SettingsByRepository(repository string) *RepositorySettings {
	if settings, ok := m.mapping[repository]; ok {
		return settings
	}
	return &m.defaultSetting
}

func SetupRepositoriesManager() (*RepositoriesManager, error) {
	repositoriesDir := config.GlobalConfig.RepositoriesDirPath
	entries, err := os.ReadDir(repositoriesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read repositories dir %s: %w", repositoriesDir, err)
	}

	manager := &RepositoriesManager{
		mapping: make(map[string]*RepositorySettings),
	}

	var defaultSettingsInitialized bool
	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			logger.Debugf("[RepositoriesSetup] directory %s skipped", name)
			continue
		}
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			logger.Debugf("[RepositoriesSetup] file %s skipped", name)
			continue
		}

		path := filepath.Join(repositoriesDir, name)
		settings, err := loadRepositorySettings(path)
		if err != nil {
			return nil, err
		}

		if settings.Repository == DefaultRepository {
			manager.defaultSetting = *settings
			defaultSettingsInitialized = true
			logger.Debugf("[RepositoriesSetup] loaded default repository settings (%s)", name)
			continue
		}

		manager.mapping[settings.Repository] = settings
		logger.Debugf("[RepositoriesSetup] loaded settings (%s) for repository %s", name, settings.Repository)
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

	if settings.Url == "" {
		settings.Url = config.GlobalConfig.GitverseBaseURL
	}

	for i, code := range settings.AllowedJiraProjects {
		settings.AllowedJiraProjects[i] = strings.ToUpper(strings.TrimSpace(code))
		// TODO: Может быть буду добавлять "-" к номеру
	}

	return &settings, nil
}
