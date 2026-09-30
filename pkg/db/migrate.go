package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const currentSchemaVersion = 1

func migrate(db *sqlx.DB) error {
	var version int
	if err := db.Get(&version, `PRAGMA user_version`); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	for version < currentSchemaVersion {
		next := version + 1
		fn, ok := migrations[next]
		if !ok {
			return fmt.Errorf("missing migration for schema version %d", next)
		}
		if err := fn(db); err != nil {
			return fmt.Errorf("migrate to version %d: %w", next, err)
		}
		if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, next)); err != nil {
			return fmt.Errorf("set schema version %d: %w", next, err)
		}
		version = next
	}

	return nil
}

var migrations = map[int]func(*sqlx.DB) error{
	1: migrateV1,
}

func migrateV1(db *sqlx.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS repositories (
	name TEXT PRIMARY KEY NOT NULL,
	allowed_jira_projects TEXT NOT NULL DEFAULT '[]',
	telegram_tags TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS action_rules (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repository_name TEXT NOT NULL REFERENCES repositories(name) ON DELETE CASCADE,
	position INTEGER NOT NULL DEFAULT 0,
	on_event TEXT NOT NULL,
	action TEXT NOT NULL,
	branch TEXT NOT NULL DEFAULT '',
	template TEXT NOT NULL DEFAULT '',
	skip_empty INTEGER NOT NULL DEFAULT 0,
	state TEXT NOT NULL DEFAULT '',
	cicd_state TEXT NOT NULL DEFAULT '',
	cicd_context TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS event_groups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repository_name TEXT NOT NULL REFERENCES repositories(name) ON DELETE CASCADE,
	position INTEGER NOT NULL DEFAULT 0,
	key TEXT NOT NULL,
	group_by TEXT NOT NULL DEFAULT '',
	strategy TEXT NOT NULL DEFAULT '',
	event TEXT NOT NULL,
	size INTEGER NOT NULL DEFAULT 0,
	ttl INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_action_rules_repository
	ON action_rules(repository_name, position);

CREATE INDEX IF NOT EXISTS idx_event_groups_repository
	ON event_groups(repository_name, position);
`
	_, err := db.Exec(schema)
	return err
}
