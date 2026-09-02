package events

type ActionType int8

const (
	JiraCommentIssue ActionType = iota
	TelegramNotify
	GitverseUpdateTitle
)

type Action struct {
	Event EventType
	Action ActionType
	// Template string
}

type ActionSettings struct {
	Repository string
	Actions []Action
}

func (s *ActionSettings) ActionsByEvent(event EventType) []Action {
	return make([]Action, 0)
}
