package actions

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/jira"
)

type JiraCommentIssue struct {
	Client *jira.JiraClient
}

func NewJiraCommentIssue(client *jira.JiraClient) *JiraCommentIssue {
	return &JiraCommentIssue{Client: client}
}

func (a *JiraCommentIssue) Name() string {
	return "jira.comment_issue"
}

// TODO: pass context to jira client, later i want to add timeout for context
func (a *JiraCommentIssue) Run(_ context.Context, ev events.Event, _ events.ActionRule) error {
	if a.Client == nil {
		return jira.ErrNotConfigured
	}

	_ = jira.IssueComment{
		Repository:  ev.Repository,
		Title:       ev.PullRequest.Title,
		Author:      ev.PullRequest.Author.Name,
		AuthorLogin: ev.PullRequest.Author.Name,
		Action:      ev.Action,
		Branch:      ev.Branch,
	}

	return nil
}
