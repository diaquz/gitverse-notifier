package dispath

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubHandler struct {
	name  string
	ready bool
	err   error

	mu   sync.Mutex
	runs []runCall
}

type runCall struct {
	event *events.Event
	rule  settings.ActionRule
}

func (h *stubHandler) Name() string { return h.name }
func (h *stubHandler) Ready() bool  { return h.ready }

func (h *stubHandler) Run(_ context.Context, ev *events.Event, rule *settings.ActionRule) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.runs = append(h.runs, runCall{event: ev, rule: *rule})
	return h.err
}

func (h *stubHandler) runCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.runs)
}

func (h *stubHandler) lastRun() (runCall, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.runs) == 0 {
		return runCall{}, false
	}
	return h.runs[len(h.runs)-1], true
}

func setupTestManager(t *testing.T, yamlContents ...string) *settings.SettingsManager {
	t.Helper()
	require.NotEmpty(t, yamlContents)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "default.yml"), []byte(yamlContents[0]), 0o644))
	for i := 1; i < len(yamlContents); i++ {
		path := filepath.Join(dir, "repo_"+string(rune('a'+i-1))+".yml")
		require.NoError(t, os.WriteFile(path, []byte(yamlContents[i]), 0o644))
	}

	prev := config.GlobalConfig
	config.GlobalConfig = &config.Config{
		RepositoriesDirPath: dir,
		GitverseBaseURL:     "https://gitverse.example",
	}
	t.Cleanup(func() { config.GlobalConfig = prev })

	manager, err := settings.SetupSettingsManager()
	require.NoError(t, err)
	return manager
}

func TestNewDispatcher(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules: []
`)

	h1 := &stubHandler{name: "utils.log", ready: true}
	h2 := &stubHandler{name: "telegram.notify", ready: true}

	d := NewDispatcher(manager, h1, h2)
	require.NotNil(t, d)
	assert.Same(t, manager, d.manager)
	require.Len(t, d.handlers, 2)
	assert.Same(t, h1, d.handlers["utils.log"])
	assert.Same(t, h2, d.handlers["telegram.notify"])
}

func TestDispatcher_Register(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules: []
`)
	d := NewDispatcher(manager)

	first := &stubHandler{name: "utils.log", ready: true}
	require.NoError(t, d.Register(first))
	assert.Same(t, first, d.handlers["utils.log"])

	replacement := &stubHandler{name: "utils.log", ready: false}
	require.NoError(t, d.Register(replacement))
	assert.Same(t, replacement, d.handlers["utils.log"])
}

func TestDispatcher_Dispatch_NoActions(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules:
  - on: pull_request.opened
    action: utils.log
`)
	handler := &stubHandler{name: "utils.log", ready: true}
	d := NewDispatcher(manager, handler)

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.BranchPush,
		Repository: "org/app",
		Branch:     "main",
	})

	assert.Equal(t, 0, handler.runCount())
}

func TestDispatcher_Dispatch_MissingHandler(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules:
  - on: pull_request.opened
    action: missing.action
`)
	other := &stubHandler{name: "utils.log", ready: true}
	d := NewDispatcher(manager, other)

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.PullRequestOpened,
		Repository: "org/app",
	})

	assert.Equal(t, 0, other.runCount())
}

func TestDispatcher_Dispatch_HandlerNotReady(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules:
  - on: pull_request.opened
    action: telegram.notify
    template: telegram/pr_opened
`)
	handler := &stubHandler{name: "telegram.notify", ready: false}
	d := NewDispatcher(manager, handler)

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.PullRequestOpened,
		Repository: "org/app",
	})

	assert.Equal(t, 0, handler.runCount())
}

func TestDispatcher_Dispatch_RunsReadyHandler(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules:
  - on: pull_request.opened
    action: utils.log
    template: unused
`)
	handler := &stubHandler{name: "utils.log", ready: true}
	d := NewDispatcher(manager, handler)

	event := &events.Event{
		Type:       events.PullRequestOpened,
		Repository: "org/app",
	}
	d.Dispatch(context.Background(), event)

	require.Equal(t, 1, handler.runCount())
	call, ok := handler.lastRun()
	require.True(t, ok)
	assert.Equal(t, event.Type, call.event.Type)
	assert.Equal(t, event.Repository, call.event.Repository)
	assert.Equal(t, "utils.log", call.rule.Action)
	assert.Equal(t, events.PullRequestOpened, call.rule.On)
}

func TestDispatcher_Dispatch_HandlerErrorDoesNotStopOthers(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules:
  - on: pull_request.opened
    action: first.action
  - on: pull_request.opened
    action: second.action
`)
	failing := &stubHandler{name: "first.action", ready: true, err: errors.New("boom")}
	ok := &stubHandler{name: "second.action", ready: true}
	d := NewDispatcher(manager, failing, ok)

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.PullRequestOpened,
		Repository: "org/app",
	})

	assert.Equal(t, 1, failing.runCount())
	assert.Equal(t, 1, ok.runCount())
}

func TestDispatcher_Dispatch_UsesRepositorySpecificRules(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules:
  - on: pull_request.opened
    action: utils.log
`, `
repository: org/app
action_rules:
  - on: pull_request.opened
    action: telegram.notify
    template: telegram/pr_opened
`)

	logHandler := &stubHandler{name: "utils.log", ready: true}
	tgHandler := &stubHandler{name: "telegram.notify", ready: true}
	d := NewDispatcher(manager, logHandler, tgHandler)

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.PullRequestOpened,
		Repository: "org/app",
	})
	assert.Equal(t, 0, logHandler.runCount())
	assert.Equal(t, 1, tgHandler.runCount())

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.PullRequestOpened,
		Repository: "org/other",
	})
	assert.Equal(t, 1, logHandler.runCount())
	assert.Equal(t, 1, tgHandler.runCount())
}

func TestDispatcher_Dispatch_RespectsBranchFilter(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules:
  - on: branch.push
    action: telegram.notify
    branch: main
`)
	handler := &stubHandler{name: "telegram.notify", ready: true}
	d := NewDispatcher(manager, handler)

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.BranchPush,
		Repository: "org/app",
		Branch:     "develop",
	})
	assert.Equal(t, 0, handler.runCount())

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.BranchPush,
		Repository: "org/app",
		Branch:     "main",
	})
	assert.Equal(t, 1, handler.runCount())
}

func TestDispatcher_Dispatch_SkipsUnknownEvent(t *testing.T) {
	manager := setupTestManager(t, `
repository: any
action_rules:
  - on: pull_request.opened
    action: utils.log
`)
	handler := &stubHandler{name: "utils.log", ready: true}
	d := NewDispatcher(manager, handler)

	d.Dispatch(context.Background(), &events.Event{
		Type:       events.Unknown,
		Repository: "org/app",
	})
	assert.Equal(t, 0, handler.runCount())
}
