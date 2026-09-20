package handlers

import (
	"context"
	"fmt"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/gitverse"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/settings"
	"gitverse-notifier/pkg/templates"
)

const defaultGitverseCommentTemplate = "gitverse/default"

type GitverseCreateComment struct {
	client    *gitverse.Client
	templates *templates.Engine
}

func NewGitverseCreateComment(client *gitverse.Client, engine *templates.Engine) *GitverseCreateComment {
	return &GitverseCreateComment{client: client, templates: engine}
}

func (a *GitverseCreateComment) Name() string {
	return "gitverse.create_comment"
}

func (a *GitverseCreateComment) Ready() bool {
	return a.client != nil
}

func (a *GitverseCreateComment) Run(ctx context.Context, event *events.Event, rule *settings.ActionRule) error {
	if event.Repository == "" || event.PullRequest.Number <= 0 {
		logger.Info(ctx, "no pull request found, skipping action",
			"action", a.Name(),
			"repository", event.Repository,
			"pr_number", event.PullRequest.Number,
		)
		return nil
	}

	templateName := rule.Template
	if templateName == "" {
		templateName = defaultGitverseCommentTemplate
	}

	body, err := a.templates.Render(templateName, templates.DataFromEvent(event))
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}
	logger.Debug(ctx, "rendered gitverse comment body", "action", a.Name(), "body", body)

	if strings.TrimSpace(body) == "" {
		logger.Info(ctx, "empty comment body, skipping action",
			"action", a.Name(),
			"template", templateName,
			"repository", event.Repository,
			"pr_number", event.PullRequest.Number,
		)
		return nil
	}

	if err := a.client.CreateIssueComment(ctx, event.Repository, event.PullRequest.Number, body); err != nil {
		return err
	}

	logger.Info(ctx, "successfully created gitverse comment",
		"action", a.Name(),
		"event", event.Type,
		"repository", event.Repository,
		"pr_number", event.PullRequest.Number,
		"template", templateName,
	)
	return nil
}
