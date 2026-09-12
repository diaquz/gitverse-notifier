package parsers

import (
	"encoding/json"
	"fmt"

	"gitverse-notifier/pkg/events"
)

type pushPayload struct {
	Ref          string `json:"ref"`
	Before       string `json:"before"`
	After        string `json:"after"`
	SHA          string `json:"sha"`
	TotalCommits int    `json:"totalCommits"`
}

func fillPushOrRef(event *events.Event, body []byte) error {
	var payload pushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse push/ref json: %w", err)
	}

	if payload.Ref != "" {
		event.Ref = payload.Ref
		event.Branch = events.BranchName(payload.Ref)
	}

	event.Push = events.PushInfo{
		Before:       payload.Before,
		After:        payload.After,
		TotalCommits: payload.TotalCommits,
	}

	return nil
}
