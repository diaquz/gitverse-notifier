package handlers

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/jira"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/settings"
)

type JiraMentionAtWeb struct {
	client *jira.JiraClient
}

func NewJiraMentionAtWeb(client *jira.JiraClient) *JiraMentionAtWeb {
	return &JiraMentionAtWeb{client: client}
}

func (a *JiraMentionAtWeb) Name() string {
	return "jira.mention_at_web"
}

func (a *JiraMentionAtWeb) Ready() bool {
	return a.client != nil
}

func (a *JiraMentionAtWeb) Run(ctx context.Context, event *events.Event, rule *settings.ActionRule) error {
	if event.PullRequest.URL == "" {
		logger.Info(ctx, "no pull request url found, skipping action",
			"action", a.Name(),
			"repository", event.Repository,
			"title", event.PullRequest.Title,
		)
		return nil
	}

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
		if _, err := a.client.AddMentionedAtWeb(key, event.PullRequest.URL, event.PullRequest.Title); err != nil {
			return err
		}
		logger.Info(ctx, "successfully added mentioned-at web link to issue",
			"action", a.Name(),
			"event", event.Type,
			"issue", key,
			"url", event.PullRequest.URL,
			"title", event.PullRequest.Title,
		)
	}

	return nil
}
