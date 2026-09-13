package events

type ActionRule struct {
	On        EventType `yaml:"-"`
	OnCode    string    `yaml:"on"`
	Action    string    `yaml:"action"`
	Branch    string    `yaml:"branch"`
	Template  string    `yaml:"template"`
	SkipEmpty bool      `yaml:"skip_empty"`
}

func (rule *ActionRule) Allowed(event *Event) (allowed bool) {
	if event == nil || rule.On != event.Type {
		return
	}
	if rule.Branch != "" && event.Branch != rule.Branch {
		return
	}
	if rule.SkipEmpty && event.Comment.Body == "" {
		return
	}

	return true
}
