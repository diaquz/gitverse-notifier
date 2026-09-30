package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/settings"
	"gitverse-notifier/pkg/store"
)

// SeedIfEmpty loads repository settings from YAML seed files when the DB has no repositories.
func SeedIfEmpty(ctx context.Context, settingsStore store.RepositorySettingsStore, seedsDir string) error {
	count, err := settingsStore.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		logger.Debug(ctx, "skip seeding, repositories already present",
			"action", "db.seed", "count", count)
		return nil
	}

	repos, err := loadSeedFiles(seedsDir)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return fmt.Errorf("database is empty and no seed files found in %s", seedsDir)
	}

	hasDefault := false
	for _, repo := range repos {
		if err := settingsStore.Upsert(ctx, repo); err != nil {
			return fmt.Errorf("seed repository %q: %w", repo.Repository, err)
		}
		if repo.Repository == settings.DefaultRepository {
			hasDefault = true
		}
		logger.Info(ctx, "seeded repository settings",
			"action", "db.seed",
			"repository", repo.Repository,
			"actions", repo.RenderActionsCodes())
	}

	if !hasDefault {
		return fmt.Errorf("seed data must include default repository %q", settings.DefaultRepository)
	}

	return nil
}

func loadSeedFiles(seedsDir string) ([]*settings.RepositorySettings, error) {
	entries, err := os.ReadDir(seedsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read seeds dir %s: %w", seedsDir, err)
	}

	repos := make([]*settings.RepositorySettings, 0)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}

		path := filepath.Join(seedsDir, name)
		repo, err := settings.LoadRepositorySettingsFile(path)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return repos, nil
}
