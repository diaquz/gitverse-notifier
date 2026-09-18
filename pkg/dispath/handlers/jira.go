package handlers

import (
	"context"
	"fmt"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/jira"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/settings"
	"gitverse-notifier/pkg/templates"
)

const defaultJiraTemplate = "jira/default"

type JiraCommentIssue struct {
	client    *jira.JiraClient
	templates *templates.Engine
}

func NewJiraCommentIssue(client *jira.JiraClient, engine *templates.Engine) *JiraCommentIssue {
	return &JiraCommentIssue{client: client, templates: engine}
}

func (a *JiraCommentIssue) Name() string {
	return "jira.comment_issue"
}

func (a *JiraCommentIssue) Ready() bool {
	return a.client != nil
}

func (a *JiraCommentIssue) Run(ctx context.Context, event events.Event, rule *settings.ActionRule) error {
	templateName := strings.TrimSpace(rule.Template)
	if templateName == "" {
		templateName = defaultJiraTemplate
	}

	body, err := a.templates.Render(templateName, templates.DataFromEvent(event))
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}
	logger.Debug(ctx, "rendered issue comment body", "body", body)
	
	if len(event.IssueKeys) == 0 {
		logger.Info(ctx, "no issue keys found, skipping action",
			"action", a.Name(),
			"repository", event.Repository,
			"title", event.PullRequest.Title,
			"branch", event.Branch,
		)
		return nil
	}

	for _, key := range event.IssueKeys {
		if _, err := a.client.AddComment(key, body); err != nil {
			return err 
		}

		logger.Info(ctx, "successfully commented on issue",
			"action", a.Name(),
			"event", event.Type,
			"issue", key,
		)
	}
	return nil
}
