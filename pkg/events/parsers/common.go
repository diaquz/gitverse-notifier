package parsers

import (
	"encoding/json"
	"fmt"

	"gitverse-notifier/pkg/events"
)

type actorPayload struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (a actorPayload) toActor() events.Actor {
	return events.Actor{ID: a.ID, Name: a.Name, Email: a.Email}
}

type repositoryPayload struct {
	FullName      string `json:"fullName"`
	FullNameSnake string `json:"full_name"`
}

func (r repositoryPayload) repositoryName() string {
	if r.FullName != "" {
		return r.FullName
	}
	return r.FullNameSnake
}

type commonPayload struct {
	Action     string            `json:"action"`
	Ref        string            `json:"ref"`
	Repository repositoryPayload `json:"repository"`
	Sender     actorPayload      `json:"sender"`
	Pusher     actorPayload      `json:"pusher"`
}

func fillCommon(event *events.Event, body []byte) error {
	if len(body) == 0 {
		return nil
	}

	var payload commonPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse event json: %w", err)
	}

	event.Action = payload.Action
	event.Ref = payload.Ref
	event.Branch = events.BranchName(payload.Ref)
	event.Repository = payload.Repository.repositoryName()
	event.Sender = payload.Sender.toActor()

	if event.Sender == (events.Actor{}) {
		event.Sender = payload.Pusher.toActor()
	}

	return nil
}
