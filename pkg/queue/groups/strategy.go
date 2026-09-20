package groups

import "gitverse-notifier/pkg/events"

type Strategy interface {
	Name() string
	Apply(group []*events.Event) []*events.Event
}

type UseLastEvent struct{}

func (UseLastEvent) Name() string { return "strategy.use_last_event" }

func (UseLastEvent) Apply(group []*events.Event) []*events.Event {
	if len(group) == 0 {
		return nil
	}

	return []*events.Event{group[len(group)-1]}
}
