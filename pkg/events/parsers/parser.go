package parsers

import (
	"context"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
)

var globalEventParser eventParser

type eventParser struct {
	enrichers []events.Enricher
}

func SetupEventParser(enrichers ...events.Enricher) error {
	globalEventParser.enrichers = enrichers
	return nil
}

func ParseEvent(ctx context.Context, eventName, eventTypeName string, body []byte) (event events.Event, err error) {
	if err := fillCommon(&event, body); err != nil {
		return event, err
	}

	event_type, err := events.ResolveEventType(eventName, eventTypeName, event.Action)
	if err != nil {
		return event, err
	}
	event.Type = event_type

	switch event.Type {
	case events.PullRequestOpened, events.PullRequestClosed, events.PullRequestEdited, events.PullRequestSynchronized,
		events.PullRequestReviewRequested, events.PullRequestReviewApproved, events.PullRequestReviewRejected,
		events.PullRequestReviewComment:
		if err := fillPullRequest(&event, body); err != nil {
			return event, err
		}
	case events.PullRequestComment:
		if err := fillIssueComment(&event, body); err != nil {
			return event, err
		}
	case events.BranchPush, events.BranchCreated, events.BranchDeleted:
		if err := fillPushOrRef(&event, body); err != nil {
			return event, err
		}
	case events.CICDStatus:
		if err := fillStatus(&event, body); err != nil {
			return event, err
		}
	}

	for _, enricher := range globalEventParser.enrichers {
		if err := enricher.Enrich(ctx, &event); err != nil {
			logger.Error(ctx, "event enriching failed",
				"action", "event_parsing",
				"enricher", enricher.Name(),
				"event", event.Type,
				"repository", event.Repository, 
				"err", err)
		}
	}

	return event, nil
}
