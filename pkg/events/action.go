package events

import "context"

type ActionRule struct {
	On        EventType
	Action    string
	Branches  []string
	Template  string
	SkipEmpty bool
}

type ActionHandler interface {
	Name() string
	Run(ctx context.Context, ev Event, rule ActionRule) error
}
