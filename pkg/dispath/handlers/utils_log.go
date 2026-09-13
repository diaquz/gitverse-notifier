package handlers

import (
	"context"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
)

type UtilsLog struct{}

func NewUtilsLog() *UtilsLog {
	return &UtilsLog{}
}

func (a *UtilsLog) Name() string {
	return "utils.log"
}

func (a *UtilsLog) Ready() bool {
	return true
}

func (a *UtilsLog) Run(ctx context.Context, ev events.Event, rule *events.ActionRule) error {
	logger.Info(ctx, "event log",
		"action", a.Name(),
		"event", ev.Type,
		"event_action", ev.Action,
		"repository", ev.Repository,
		"branch", ev.Branch,
		"ref", ev.Ref,
		"sender", ev.Sender.Name,
		"issue_keys", strings.Join(ev.IssueKeys, ","),
		"pr", ev.PullRequest.Number,
		"title", ev.PullRequest.Title,
		"comment", ev.Comment.Body,
		"status_context", ev.Status.Context,
		"status_state", ev.Status.State,
		"template", rule.Template,
	)
	return nil
}
