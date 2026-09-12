package events

import "context"

type ActionRule struct {
	On        EventType `yaml:"-"`
	OnCode    string    `yaml:"on"`
	Action    string    `yaml:"action"`
	Branch    string    `yaml:"branch"`
	Template  string    `yaml:"template"`
	SkipEmpty bool      `yaml:"skip_empty"`
}

type ActionHandler interface {
	Name() string
	Run(ctx context.Context, ev Event, rule ActionRule) error
}
