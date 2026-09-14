package events

import "strings"

type ActionRule struct {
	On          EventType `yaml:"-"`
	OnCode      string    `yaml:"on"`
	Action      string    `yaml:"action"`
	Branch      string    `yaml:"branch"`
	Template    string    `yaml:"template"`
	SkipEmpty   bool      `yaml:"skip_empty"`
	CICDState   string    `yaml:"cicd_state"`
	CICDContext string    `yaml:"cicd_context"`
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

func contains(str, subStr string) bool {
	return strings.Contains(
		strings.ToLower(str),
		strings.ToLower(subStr),
	)
}
