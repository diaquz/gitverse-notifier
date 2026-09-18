package batches

import "gitverse-notifier/pkg/events"

type Strategy interface {
	Name() string
	Apply(batch []*events.Event) []*events.Event
}

type UseLastEvent struct{}

func (UseLastEvent) Name() string { return "use_last_event" }

func (UseLastEvent) Apply(batch []*events.Event) []*events.Event {
	if len(batch) == 0 {
		return nil
	}

	return []*events.Event{batch[len(batch)-1]}
}
