package events

type Enricher interface {
	Name() string
	Enrich(event *Event) error
}
