package parsers

import (
	"context"
	"gitverse-notifier/pkg/events"
)

func ParseEvent(_ context.Context, eventName, eventTypeName, requestId string, body []byte) (event events.Event, err error) {
	if err := fillCommon(&event, body); err != nil {
		return event, err
	}

	event_type, err := events.ResolveEventType(eventName, eventTypeName, event.Action)
	if err != nil {
		return event, err
	}
	event.Type = event_type
	event.RequestId = requestId

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

	return event, nil
}
