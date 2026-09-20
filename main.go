package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitverse-notifier/pkg/cache"
	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/dispatch"
	"gitverse-notifier/pkg/dispatch/handlers"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/events/enrichers"
	"gitverse-notifier/pkg/integrations/gitverse"
	"gitverse-notifier/pkg/integrations/jira"
	"gitverse-notifier/pkg/integrations/telegram"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/pipeline"
	gvqueries "gitverse-notifier/pkg/queries/gitverse"
	"gitverse-notifier/pkg/server"
	"gitverse-notifier/pkg/settings"
	"gitverse-notifier/pkg/templates"
)

var (
	configPath = ""
)

func init() {
	flag.StringVar(&configPath, "f", "", "config.yml path")
}

func main() {
	flag.Parse()
	config.Setup(configPath)
	logger.SetupLogger(config.GlobalConfig)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	p := buildEventPipeline(ctx)
	p.StartWorkers(ctx)

	s := server.NewHttpServer(p)
	go runHttpServer(ctx, s)

	<-ctx.Done()
	logger.Info(ctx, "shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "http server shutdown failed", "err", err)
	}

	if err := p.Close(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "pipeline shutdown failed", "err", err)
	}

	logger.Info(shutdownCtx, "shutdown complete")
}

func buildEventPipeline(ctx context.Context) *pipeline.Pipeline {
	manager, err := settings.SetupSettingsManager()
	if err != nil {
		logger.Fatal(ctx, err)
	}

	engine, err := templates.SetupTemplateEngine()
	if err != nil {
		logger.Fatal(ctx, err)
	}

	jiraClient, jiraErr := jira.NewJiraClient()
	if jiraErr != nil {
		logger.Fatal(ctx, jiraErr)
	}

	tgClient, tgErr := telegram.NewTelegramClient()
	if tgErr != nil {
		logger.Error(ctx, "failed to configure telegram client", "err", tgErr)
	}

	dispatcher := dispatch.NewDispatcher(
		manager,
		handlers.NewJiraCommentIssue(jiraClient, engine),
		handlers.NewJiraMentionAtWeb(jiraClient),
		handlers.NewTelegramNotify(tgClient, engine),
		handlers.NewUtilsLog(),
	)

	gitverseClient, gitverseErr := gitverse.NewClient()
	if gitverseErr != nil {
		logger.Error(ctx, "failed to configure gitverse client", "err", gitverseErr)
	}

	eventEnrichers := make([]events.Enricher, 0, 5)
	if gitverseClient != nil {
		queries := gvqueries.New(gitverseClient, cache.NewMemoryPullRequestCache())
		eventEnrichers = append(eventEnrichers, enrichers.NewGitversePullRequest(queries))
		eventEnrichers = append(eventEnrichers, enrichers.NewGitverseCommit(queries))
	}

	eventEnrichers = append(eventEnrichers, enrichers.NewJiraIssueKeys(manager))
	eventEnrichers = append(eventEnrichers, enrichers.NewGitverseLinks(manager))
	eventEnrichers = append(eventEnrichers, enrichers.NewTelegramLinks(manager))

	return pipeline.BuildNewPipeline(
		manager,
		eventEnrichers,
		dispatcher,
	)
}

func runHttpServer(ctx context.Context, s *server.HttpServer) {
	if err := s.Run(); err != nil {
		logger.Fatal(ctx, err)
	}
}
