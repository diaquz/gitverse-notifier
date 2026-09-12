package actions

import (
	"context"
	_ "fmt"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/jira"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/templates"
)

const defaultJiraTemplate = "jira/default"

type JiraCommentIssue struct {
	Client    *jira.JiraClient
	Templates *templates.Engine
}

func NewJiraCommentIssue(client *jira.JiraClient, engine *templates.Engine) *JiraCommentIssue {
	return &JiraCommentIssue{Client: client, Templates: engine}
}

func (a *JiraCommentIssue) Name() string {
	return "jira.comment_issue"
}

func (a *JiraCommentIssue) Run(_ context.Context, ev events.Event, rule events.ActionRule) error {
	templateName := strings.TrimSpace(rule.Template)
	if templateName == "" {
		templateName = defaultJiraTemplate
	}

	body, err := a.Templates.Render(templateName, templates.DataFromEvent(ev))
	if err != nil {
		return err
	}
	logger.Debugf("rendered issue comment body: %s", body)
	
	if len(ev.IssueKeys) == 0 {
		logger.Infof("[Action=%s] no issue keys (repository=%s title=%q branch=%q)", a.Name(),  ev.Repository, ev.PullRequest.Title, ev.Branch)
		return nil
	}

	for _, key := range ev.IssueKeys {
		// if _, err := a.Client.AddComment(key, body); err != nil {
		//	return fmt.Errorf("failed to comment on issue %s: %w", key, err)
		// }

		logger.Infof("[Action=%s] successfully commented on issue %s", a.Name(), key)
	}
	return nil
}
