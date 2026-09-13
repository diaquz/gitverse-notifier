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

func (a *UtilsLog) Run(_ context.Context, ev events.Event, rule *events.ActionRule) error {
	logger.Infof(
		"[Action=%s] event=%s action=%q repository=%q branch=%q ref=%q sender=%q issue_keys=%q pr=%d title=%q comment=%q status=%q/%q template=%q",
		a.Name(),
		ev.Type,
		ev.Action,
		ev.Repository,
		ev.Branch,
		ev.Ref,
		ev.Sender.Name,
		strings.Join(ev.IssueKeys, ","),
		ev.PullRequest.Number,
		ev.PullRequest.Title,
		ev.Comment.Body,
		ev.Status.Context,
		ev.Status.State,
		rule.Template,
	)

	return nil
}
