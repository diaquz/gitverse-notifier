package events

import "context"

type Enricher interface {
	Name() string
	Enrich(ctx context.Context, event *Event) error
	Skip(event *Event) bool
}
