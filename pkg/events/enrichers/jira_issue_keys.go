package enrichers

import (
	"context"
	"regexp"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/settings"
)

var (
	issueKeyRe = regexp.MustCompile(`\b([A-Z][A-Z0-9]+-\d+)\b`)
)

type JiraIssueKeys struct {
	manager *settings.SettingsManager
}

func NewJiraIssueKeys(manager *settings.SettingsManager) *JiraIssueKeys {
	return &JiraIssueKeys{manager: manager}
}

func (e *JiraIssueKeys) Name() string {
	return "enricher.jira-issue-keys"
}

func (e *JiraIssueKeys) Skip(event *events.Event) bool {
	return event == nil
}

func (e *JiraIssueKeys) Enrich(ctx context.Context, event *events.Event) error {
	keys := make([]string, 0, 1)
	seen := make(map[string]struct{})

	add := func(key string) {
		key = strings.ToUpper(strings.TrimSpace(key))
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}

		if !e.manager.IsJiraCodeAllowed(event.Repository, key) {
			logger.Debug(ctx, "jira issue key is permited",
				"action", e.Name(),
				"issue-key", key,
				"event", event.Type,
				"repository", event.Repository)
			return
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	for _, key := range issueKeysInText(event.PullRequest.Title) {
		add(key)
	}

	for _, key := range issueKeysInText(event.PullRequest.Body) {
		add(key)
	}

	for _, key := range issueKeysInText(event.Push.CommitTitle) {
		add(key)
	}

	for _, key := range issueKeysInText(event.Branch) {
		add(key)
	}

	event.IssueKeys = keys
	return nil
}

// issueKeysInText возвращает номера задач jira, которые содержатся в тексте
//
//	JIRA-1
//	JIRA-2 JIRA-3 Заголовок
//	Слияние из ветки JIRA-1 в develop
func issueKeysInText(text string) []string {
	if text == "" {
		return nil
	}

	matches := issueKeyRe.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	return matches
}
