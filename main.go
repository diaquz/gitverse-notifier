package main

import (
	"context"
	"flag"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/dispath"
	"gitverse-notifier/pkg/dispath/handlers"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/events/enrichers"
	"gitverse-notifier/pkg/integrations/gitverse"
	"gitverse-notifier/pkg/integrations/jira"
	"gitverse-notifier/pkg/integrations/telegram"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/repositories"
	"gitverse-notifier/pkg/server"
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
	ctx := context.Background()

	manager, err := repositories.SetupRepositoriesManager()
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

	dispatcher := dispath.NewDispatcher(
		manager,
		handlers.NewJiraCommentIssue(jiraClient, engine),
		handlers.NewTelegramNotify(tgClient, engine),
		handlers.NewUtilsLog(),
	)

	eventEnrichers := make([]events.Enricher, 0, 3)

	gitverseClient, gitverseErr := gitverse.NewClient()
	if gitverseErr != nil {
		logger.Error(ctx, "failed to configure gitverse client", "err", gitverseErr)
	} else {
		eventEnrichers = append(eventEnrichers,
			enrichers.NewGitversePullRequest(gitverseClient))
	}

	eventEnrichers = append(eventEnrichers,
		enrichers.NewJiraIssueKeys(manager),
		enrichers.NewGitverseLinks(manager),
	)

	proc := processor.New(dispatcher, eventEnrichers...)
	srv := server.NewHttpServer(proc)
	logger.Fatal(ctx, srv.Run())
}
