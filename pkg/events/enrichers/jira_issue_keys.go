package enrichers

import (
	"regexp"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/repositories"
)

var (
	issueKeyRe        = regexp.MustCompile(`\b([A-Z][A-Z0-9]+-\d+)\b`)
	leadingIssueKeyRe = regexp.MustCompile(`^([A-Z][A-Z0-9]+-\d+)\b`)
)

type JiraIssueKeys struct {
	manager *repositories.RepositoriesManager
}

func NewJiraIssueKeys(manager *repositories.RepositoriesManager) *JiraIssueKeys {
	return &JiraIssueKeys{manager: manager}
}

func (e *JiraIssueKeys) Name() string {
	return "jira.issue_keys"
}

func (e *JiraIssueKeys) Enrich(event *events.Event) error {
	if event == nil {
		return nil
	}

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
			logger.Debugf("[JiraCodesEnricher] code %s is permited for repository %s", key, event.Repository)
			return
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	for _, key := range leadingIssueKeys(event.PullRequest.Title) {
		add(key)
	}

	for _, key := range issueKeysInRef(event.Branch) {
		add(key)
	}

	event.IssueKeys = keys
	return nil
}

// leadingIssueKeys возвращает кода задач jira, содержащиеся в заголовке PR, например
//
//	JIRA-1 Заголовок
//	JIRA-2 JIRA-3 Заголовок
func leadingIssueKeys(title string) []string {
	rest := strings.TrimSpace(title)
	var out []string

	for rest != "" {
		m := leadingIssueKeyRe.FindStringSubmatch(rest)
		if m == nil {
			break
		}

		out = append(out, m[1])
		rest = strings.TrimSpace(rest[len(m[0]):])
	}
	return out
}

// issueKeysInRef возвращает номера задач jira, которые содержатся в имени ветки, например
//
//	JIRA-1
func issueKeysInRef(text string) []string {
	matches := issueKeyRe.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	return matches
}
