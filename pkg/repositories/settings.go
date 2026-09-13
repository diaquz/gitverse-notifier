package repositories

import (
	"gitverse-notifier/pkg/events"
	"strings"
)

// RepositorySettings хранит настройки конкретного репозитория
type RepositorySettings struct {
	Repository          string              `yaml:"repository"`
	Url                 string              `yaml:"url"`
	AllowedJiraProjects []string            `yaml:"allowed_jira_projects"`
	Actions             []events.ActionRule `yaml:"action_rules"`
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

func (s *RepositorySettings) ActionsByEvent(event events.Event) []events.ActionRule {
	actions := make([]events.ActionRule, 0)
	for _, rule := range s.Actions {
		if !rule.Allowed(&event) {
			continue
		}

		actions = append(actions, rule)
	}
	return actions
}

func (s *RepositorySettings) RenderActionsCodes() string {
	codes := make([]string, 0, len(s.Actions))
	for _, action := range s.Actions {
		codes = append(codes, action.Action)
	}

	return strings.Join(codes, ", ")
}
