package events

import "strings"

type ActionRule struct {
	On          EventType `yaml:"-"`
	OnCode      string    `yaml:"on"`
	Action      string    `yaml:"action"`
	Branch      string    `yaml:"branch"`
	Template    string    `yaml:"template"`
	SkipEmpty   bool      `yaml:"skip_empty"`
	State       string    `yaml:"state"`
	CICDState   string    `yaml:"cicd_state"`
	CICDContext string    `yaml:"cicd_context"`
}

func (rule *ActionRule) Allowed(event *Event) (allowed bool) {
	if !rule.matchesType(event) {
		return
	}

	if rule.SkipEmpty && event.Comment.Body == "" {
		return
	}

	if rule.Branch != "" && event.Branch != rule.Branch {
		return
	}

	if !rule.matchesState(event) {
		return
	}

	if event.Type == CICDStatus && !rule.cicdStatusAllowed(event) {
		return
	}

	return true
}

func (rule *ActionRule) cicdStatusAllowed(event *Event) (allowed bool) {
	if rule.CICDContext != "" && !contains(event.Status.Context, rule.CICDContext) {
		return
	}

	if rule.CICDState != "" && event.Status.State != rule.CICDState {
		return
	}

	return true
}

// PotentiallyAllowed проверяет только тип
func (rule *ActionRule) PotentiallyAllowed(event *Event) bool {
	return rule.matchesType(event)
}

func (rule *ActionRule) matchesType(event *Event) bool {
	if event == nil || rule.On != event.Type {
		return false
	}

	return true
}

func contains(str, subStr string) bool {
	return strings.Contains(
		strings.ToLower(str),
		strings.ToLower(subStr),
	)
}

func (rule *ActionRule) matchesState(event *Event) bool {
	if rule.State == "" {
		return true
	}

	switch event.Type {
	case PullRequestClosed:
		return rule.State == event.PullRequest.EffectiveState()
	case CICDStatus:
		return rule.State == event.Status.State
	default:
		return false
	}
}
