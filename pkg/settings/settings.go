package settings

import (
	"fmt"
	"gitverse-notifier/pkg/events"
	"strings"
	"time"
)

type EventGroupSettings struct {
	Key      string           `yaml:"key"`
	GroupBy  string           `yaml:"group_by"`
	Strategy string           `yaml:"strategy"`
	EventRaw string           `yaml:"event"`
	Event    events.EventType `yaml:"-"`
	Size     int              `yaml:"size"`
	TTL      int              `yaml:"ttl"`
}

func (b *EventGroupSettings) TTLTime() time.Duration {
	return time.Duration(b.TTL) * time.Second
}

// RepositorySettings хранит настройки конкретного репозитория
type RepositorySettings struct {
	Repository          string               `yaml:"repository"`
	AllowedJiraProjects []string             `yaml:"allowed_jira_projects"`
	Actions             []ActionRule         `yaml:"action_rules"`
	Groups              []EventGroupSettings `yaml:"event_groups"`
	TelegramTags        map[string]string    `yaml:"telegram_tags"`
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

func (s *RepositorySettings) ActionsByEvent(event *events.Event) []ActionRule {
	actions := make([]ActionRule, 0)
	for _, rule := range s.Actions {
		if !rule.Allowed(event) {
			continue
		}

		actions = append(actions, rule)
	}
	return actions
}

func (s *RepositorySettings) HasPotentialActions(event *events.Event) bool {
	for _, rule := range s.Actions {
		if rule.PotentiallyAllowed(event) {
			return true
		}
	}
	return false
}

func (s *RepositorySettings) RenderActionsCodes() string {
	codes := make([]string, 0, len(s.Actions))
	for _, action := range s.Actions {
		codes = append(codes,
			fmt.Sprintf("%s: %s", action.On, action.Action))
	}

	return strings.Join(codes, ", ")
}

func (s *RepositorySettings) FindEventGroupSettings(event *events.Event) *EventGroupSettings {
	for i := range len(s.Groups) {
		if s.Groups[i].Event == event.Type {
			return &s.Groups[i]
		}
	}

	return nil
}
