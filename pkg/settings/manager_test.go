package settings

import (
	"os"
	"path/filepath"
	"testing"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(defaultSettings RepositorySettings, repos map[string]*RepositorySettings) *SettingsManager {
	if repos == nil {
		repos = make(map[string]*RepositorySettings)
	}
	return &SettingsManager{
		mapping:        repos,
		defaultSetting: defaultSettings,
	}
}

func TestSettingsByRepository(t *testing.T) {
	t.Parallel()

	specific := &RepositorySettings{
		Repository:          "org/app",
		AllowedJiraProjects: []string{"APP"},
	}
	manager := newTestManager(
		RepositorySettings{
			Repository:          DefaultRepository,
			AllowedJiraProjects: []string{"JIRA"},
		},
		map[string]*RepositorySettings{
			"org/app": specific,
		},
	)

	t.Run("returns specific settings when repository is mapped", func(t *testing.T) {
		t.Parallel()

		got := manager.SettingsByRepository("org/app")
		require.NotNil(t, got)
		assert.Same(t, specific, got)
	})

	t.Run("returns default settings for unknown repository", func(t *testing.T) {
		t.Parallel()

		got := manager.SettingsByRepository("org/unknown")
		require.NotNil(t, got)
		assert.Equal(t, DefaultRepository, got.Repository)
	})
}

func TestIsJiraCodeAllowed(t *testing.T) {
	t.Parallel()

	manager := newTestManager(
		RepositorySettings{
			Repository:          DefaultRepository,
			AllowedJiraProjects: []string{"JIRA", "TEST"},
		},
		map[string]*RepositorySettings{
			"org/app": {
				Repository:          "org/app",
				AllowedJiraProjects: []string{"APP"},
			},
			"org/open": {
				Repository: "org/open",
			},
		},
	)

	tests := []struct {
		name       string
		repository string
		code       string
		want       bool
	}{
		{
			name:       "allowed for specific repository",
			repository: "org/app",
			code:       "APP-42",
			want:       true,
		},
		{
			name:       "denied for specific repository",
			repository: "org/app",
			code:       "JIRA-1",
			want:       false,
		},
		{
			name:       "falls back to default repository rules",
			repository: "org/missing",
			code:       "TEST-7",
			want:       true,
		},
		{
			name:       "empty allow-list permits any code",
			repository: "org/open",
			code:       "ANYTHING-1",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, manager.IsJiraCodeAllowed(tt.repository, tt.code))
		})
	}
}

func TestActionsFor(t *testing.T) {
	t.Parallel()

	openedRule := ActionRule{
		On:       events.PullRequestOpened,
		OnCode:   string(events.PullRequestOpened),
		Action:   "telegram.notify",
		Template: "telegram/pr_opened",
	}
	pushMainRule := ActionRule{
		On:       events.BranchPush,
		OnCode:   string(events.BranchPush),
		Action:   "telegram.notify",
		Branch:   "main",
		Template: "telegram/branch_push",
	}

	manager := newTestManager(
		RepositorySettings{
			Repository: DefaultRepository,
			Actions:    []ActionRule{openedRule},
		},
		map[string]*RepositorySettings{
			"org/app": {
				Repository: "org/app",
				Actions:    []ActionRule{openedRule, pushMainRule},
			},
		},
	)

	t.Run("returns matched actions for repository event", func(t *testing.T) {
		t.Parallel()

		got := manager.ActionsFor("org/app", &events.Event{
			Type:   events.PullRequestOpened,
			Branch: "feature",
		})
		require.Len(t, got, 1)
		assert.Equal(t, "telegram.notify", got[0].Action)
		assert.Equal(t, events.PullRequestOpened, got[0].On)
	})

	t.Run("filters actions by branch", func(t *testing.T) {
		t.Parallel()

		matched := manager.ActionsFor("org/app", &events.Event{
			Type:   events.BranchPush,
			Branch: "main",
		})
		require.Len(t, matched, 1)
		assert.Equal(t, "main", matched[0].Branch)

		unmatched := manager.ActionsFor("org/app", &events.Event{
			Type:   events.BranchPush,
			Branch: "develop",
		})
		assert.Empty(t, unmatched)
	})

	t.Run("uses default repository actions for unknown repo", func(t *testing.T) {
		t.Parallel()

		got := manager.ActionsFor("org/missing", &events.Event{Type: events.PullRequestOpened})
		require.Len(t, got, 1)
		assert.Equal(t, openedRule.Action, got[0].Action)
	})

	t.Run("returns nil for unknown event type", func(t *testing.T) {
		t.Parallel()

		got := manager.ActionsFor("org/app", &events.Event{Type: events.Unknown})
		assert.Nil(t, got)
	})
}

func TestLoadRepositorySettings(t *testing.T) {
	prev := config.GlobalConfig
	config.GlobalConfig = &config.Config{GitverseBaseURL: "https://gitverse.example"}
	t.Cleanup(func() { config.GlobalConfig = prev })

	t.Run("loads settings and normalizes fields", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "app.yml")
		content := `
repository: org/app
url: ""
allowed_jira_projects:
  - " app "
  - "jira"
action_rules:
  - on: pull_request.opened
    action: telegram.notify
    template: telegram/pr_opened
`
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		got, err := loadRepositorySettings(path)
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, "org/app", got.Repository)
		assert.Equal(t, []string{"APP", "JIRA"}, got.AllowedJiraProjects)
		require.Len(t, got.Actions, 1)
		assert.Equal(t, events.PullRequestOpened, got.Actions[0].On)
		assert.Equal(t, "telegram.notify", got.Actions[0].Action)
	})

	t.Run("requires repository name", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "broken.yml")
		require.NoError(t, os.WriteFile(path, []byte("url: https://example\n"), 0o644))

		got, err := loadRepositorySettings(path)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repository name is required")
	})

	t.Run("rejects unknown action event", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "broken.yml")
		content := `
repository: org/app
action_rules:
  - on: not.a.real.event
    action: utils.log
`
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		got, err := loadRepositorySettings(path)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown event")
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		got, err := loadRepositorySettings(filepath.Join(t.TempDir(), "missing.yml"))
		assert.Nil(t, got)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read")
	})

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "broken.yml")
		require.NoError(t, os.WriteFile(path, []byte(":\n  - broken"), 0o644))

		got, err := loadRepositorySettings(path)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})
}

func TestSetupSettingsManager(t *testing.T) {
	prev := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = prev })

	t.Run("loads default and specific settings", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "default.yml"), []byte(`
repository: any
allowed_jira_projects:
  - JIRA
action_rules:
  - on: pull_request.opened
    action: utils.log
`), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "app.yaml"), []byte(`
repository: org/app
allowed_jira_projects:
  - APP
action_rules:
  - on: branch.push
    action: telegram.notify
    branch: main
`), 0o644))
		require.NoError(t, os.Mkdir(filepath.Join(dir, "subdir"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore me"), 0o644))

		config.GlobalConfig = &config.Config{
			ConfigsDirPath: dir,
			GitverseBaseURL:     "https://gitverse.example",
		}

		manager, err := SetupSettingsManager()
		require.NoError(t, err)
		require.NotNil(t, manager)

		assert.Equal(t, DefaultRepository, manager.defaultSetting.Repository)
		assert.Equal(t, []string{"JIRA"}, manager.defaultSetting.AllowedJiraProjects)
		require.Contains(t, manager.mapping, "org/app")
		assert.Equal(t, []string{"APP"}, manager.mapping["org/app"].AllowedJiraProjects)

		assert.True(t, manager.IsJiraCodeAllowed("org/app", "APP-1"))
		assert.False(t, manager.IsJiraCodeAllowed("org/app", "JIRA-1"))
		assert.True(t, manager.IsJiraCodeAllowed("org/other", "JIRA-1"))

		actions := manager.ActionsFor("org/app", &events.Event{Type: events.BranchPush, Branch: "main"})
		require.Len(t, actions, 1)
		assert.Equal(t, "telegram.notify", actions[0].Action)
	})

	t.Run("fails when default repository is missing", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "app.yml"), []byte(`
repository: org/app
action_rules: []
`), 0o644))

		config.GlobalConfig = &config.Config{ConfigsDirPath: dir}

		manager, err := SetupSettingsManager()
		assert.Nil(t, manager)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "default repository config")
	})

	t.Run("fails when settings directory is missing", func(t *testing.T) {
		config.GlobalConfig = &config.Config{
			ConfigsDirPath: filepath.Join(t.TempDir(), "does-not-exist"),
		}

		manager, err := SetupSettingsManager()
		assert.Nil(t, manager)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read settings dir")
	})

	t.Run("fails when a repository file is invalid", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "default.yml"), []byte(`
repository: any
action_rules: []
`), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.yml"), []byte(`
repository: org/broken
action_rules:
  - on: bad.event
    action: utils.log
`), 0o644))

		config.GlobalConfig = &config.Config{ConfigsDirPath: dir}

		manager, err := SetupSettingsManager()
		assert.Nil(t, manager)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown event")
	})
}
