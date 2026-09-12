package parsers

import (
	"encoding/json"
	"fmt"

	"gitverse-notifier/pkg/events"
)

type statusPayload struct {
	State       string `json:"state"`
	Context     string `json:"context"`
	Description string `json:"description"`
	SHA         string `json:"sha"`
}

func fillStatus(event *events.Event, body []byte) error {
	var payload statusPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse status json: %w", err)
	}

	event.Status = events.StatusInfo{
		State:       payload.State,
		Context:     payload.Context,
		Description: payload.Description,
		SHA:         payload.SHA,
	}
	return nil
}
