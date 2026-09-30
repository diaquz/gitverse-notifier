package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/settings"

	"github.com/jmoiron/sqlx"
)

type RepositorySettingsStore interface {
	Count(ctx context.Context) (int, error)
	List(ctx context.Context) ([]*settings.RepositorySettings, error)
	Get(ctx context.Context, name string) (*settings.RepositorySettings, error)
	Upsert(ctx context.Context, repo *settings.RepositorySettings) error
	Delete(ctx context.Context, name string) error
}

type SQLRepositorySettingsStore struct {
	db *sqlx.DB
}

func NewRepositorySettingsStore(db *sqlx.DB) *SQLRepositorySettingsStore {
	return &SQLRepositorySettingsStore{db: db}
}

type repositoryRow struct {
	Name                 string `db:"name"`
	AllowedJiraProjects  string `db:"allowed_jira_projects"`
	TelegramTags         string `db:"telegram_tags"`
}

type actionRuleRow struct {
	ID             int64  `db:"id"`
	RepositoryName string `db:"repository_name"`
	Position       int    `db:"position"`
	OnEvent        string `db:"on_event"`
	Action         string `db:"action"`
	Branch         string `db:"branch"`
	Template       string `db:"template"`
	SkipEmpty      bool   `db:"skip_empty"`
	State          string `db:"state"`
	CICDState      string `db:"cicd_state"`
	CICDContext    string `db:"cicd_context"`
}

type eventGroupRow struct {
	ID             int64  `db:"id"`
	RepositoryName string `db:"repository_name"`
	Position       int    `db:"position"`
	Key            string `db:"key"`
	GroupBy        string `db:"group_by"`
	Strategy       string `db:"strategy"`
	Event          string `db:"event"`
	Size           int    `db:"size"`
	TTL            int    `db:"ttl"`
}

func (s *SQLRepositorySettingsStore) Count(ctx context.Context) (int, error) {
	var count int
	if err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM repositories`); err != nil {
		return 0, fmt.Errorf("count repositories: %w", err)
	}
	return count, nil
}

func (s *SQLRepositorySettingsStore) List(ctx context.Context) ([]*settings.RepositorySettings, error) {
	var rows []repositoryRow
	if err := s.db.SelectContext(ctx, &rows, `SELECT name, allowed_jira_projects, telegram_tags FROM repositories ORDER BY name`); err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}

	result := make([]*settings.RepositorySettings, 0, len(rows))
	for i := range rows {
		repo, err := s.loadRepository(ctx, s.db, rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, repo)
	}
	return result, nil
}

func (s *SQLRepositorySettingsStore) Get(ctx context.Context, name string) (*settings.RepositorySettings, error) {
	var row repositoryRow
	err := s.db.GetContext(ctx, &row,
		`SELECT name, allowed_jira_projects, telegram_tags FROM repositories WHERE name = ?`, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get repository %q: %w", name, err)
	}
	return s.loadRepository(ctx, s.db, row)
}

func (s *SQLRepositorySettingsStore) Upsert(ctx context.Context, repo *settings.RepositorySettings) error {
	if repo == nil {
		return fmt.Errorf("repository settings is nil")
	}
	if repo.Repository == "" {
		return fmt.Errorf("repository name is required")
	}

	normalized, err := settings.NormalizeRepositorySettings(*repo)
	if err != nil {
		return err
	}

	if normalized.AllowedJiraProjects == nil {
		normalized.AllowedJiraProjects = []string{}
	}
	if normalized.TelegramTags == nil {
		normalized.TelegramTags = map[string]string{}
	}

	projectsJSON, err := json.Marshal(normalized.AllowedJiraProjects)
	if err != nil {
		return fmt.Errorf("marshal allowed_jira_projects: %w", err)
	}
	tagsJSON, err := json.Marshal(normalized.TelegramTags)
	if err != nil {
		return fmt.Errorf("marshal telegram_tags: %w", err)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin upsert: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO repositories (name, allowed_jira_projects, telegram_tags)
VALUES (?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
	allowed_jira_projects = excluded.allowed_jira_projects,
	telegram_tags = excluded.telegram_tags
`, normalized.Repository, string(projectsJSON), string(tagsJSON)); err != nil {
		return fmt.Errorf("upsert repository %q: %w", normalized.Repository, err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM action_rules WHERE repository_name = ?`, normalized.Repository); err != nil {
		return fmt.Errorf("clear action_rules for %q: %w", normalized.Repository, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM event_groups WHERE repository_name = ?`, normalized.Repository); err != nil {
		return fmt.Errorf("clear event_groups for %q: %w", normalized.Repository, err)
	}

	for i, rule := range normalized.Actions {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO action_rules (
	repository_name, position, on_event, action, branch, template,
	skip_empty, state, cicd_state, cicd_context
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, normalized.Repository, i, rule.OnCode, rule.Action, rule.Branch, rule.Template,
			rule.SkipEmpty, rule.State, rule.CICDState, rule.CICDContext); err != nil {
			return fmt.Errorf("insert action_rule for %q: %w", normalized.Repository, err)
		}
	}

	for i, group := range normalized.Groups {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO event_groups (
	repository_name, position, key, group_by, strategy, event, size, ttl
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, normalized.Repository, i, group.Key, group.GroupBy, group.Strategy, group.EventRaw, group.Size, group.TTL); err != nil {
			return fmt.Errorf("insert event_group for %q: %w", normalized.Repository, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit upsert for %q: %w", normalized.Repository, err)
	}
	return nil
}

func (s *SQLRepositorySettingsStore) Delete(ctx context.Context, name string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM repositories WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("delete repository %q: %w", name, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("repository %q not found", name)
	}
	return nil
}

func (s *SQLRepositorySettingsStore) loadRepository(ctx context.Context, q sqlx.QueryerContext, row repositoryRow) (*settings.RepositorySettings, error) {
	repo := &settings.RepositorySettings{
		Repository: row.Name,
	}

	if err := json.Unmarshal([]byte(row.AllowedJiraProjects), &repo.AllowedJiraProjects); err != nil {
		return nil, fmt.Errorf("parse allowed_jira_projects for %q: %w", row.Name, err)
	}
	if err := json.Unmarshal([]byte(row.TelegramTags), &repo.TelegramTags); err != nil {
		return nil, fmt.Errorf("parse telegram_tags for %q: %w", row.Name, err)
	}

	var actionRows []actionRuleRow
	if err := sqlx.SelectContext(ctx, q, &actionRows, `
SELECT id, repository_name, position, on_event, action, branch, template,
	skip_empty, state, cicd_state, cicd_context
FROM action_rules
WHERE repository_name = ?
ORDER BY position, id
`, row.Name); err != nil {
		return nil, fmt.Errorf("list action_rules for %q: %w", row.Name, err)
	}

	repo.Actions = make([]settings.ActionRule, 0, len(actionRows))
	for _, ar := range actionRows {
		eventType, ok := events.ParseEventType(ar.OnEvent)
		if !ok {
			return nil, fmt.Errorf("unknown action event %q for repository %q", ar.OnEvent, row.Name)
		}
		repo.Actions = append(repo.Actions, settings.ActionRule{
			On:          eventType,
			OnCode:      ar.OnEvent,
			Action:      ar.Action,
			Branch:      ar.Branch,
			Template:    ar.Template,
			SkipEmpty:   ar.SkipEmpty,
			State:       ar.State,
			CICDState:   ar.CICDState,
			CICDContext: ar.CICDContext,
		})
	}

	var groupRows []eventGroupRow
	if err := sqlx.SelectContext(ctx, q, &groupRows, `
SELECT id, repository_name, position, key, group_by, strategy, event, size, ttl
FROM event_groups
WHERE repository_name = ?
ORDER BY position, id
`, row.Name); err != nil {
		return nil, fmt.Errorf("list event_groups for %q: %w", row.Name, err)
	}

	repo.Groups = make([]settings.EventGroupSettings, 0, len(groupRows))
	for _, gr := range groupRows {
		eventType, ok := events.ParseEventType(gr.Event)
		if !ok {
			return nil, fmt.Errorf("unknown group event %q for repository %q", gr.Event, row.Name)
		}
		repo.Groups = append(repo.Groups, settings.EventGroupSettings{
			Key:      gr.Key,
			GroupBy:  gr.GroupBy,
			Strategy: gr.Strategy,
			EventRaw: gr.Event,
			Event:    eventType,
			Size:     gr.Size,
			TTL:      gr.TTL,
		})
	}

	return repo, nil
}
