package main

import (
	"flag"

	"gitverse-notifier/pkg/dispath/handlers"
	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/dispath"
	"gitverse-notifier/pkg/events/enrichers"
	"gitverse-notifier/pkg/events/parsers"
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

	manager, err := repositories.SetupRepositoriesManager()
	if err != nil {
		logger.Fatal(err)
	}

	engine, err := templates.SetupTemplateEngine()
	if err != nil {
		logger.Fatal(err)
	}

	if err := parsers.SetupEventParser(
		enrichers.NewJiraIssueKeys(manager),
		enrichers.NewGitverseLinks(manager),
	); err != nil {
		logger.Fatal(err)
	}

	jiraClient, jiraErr := jira.NewJiraClient()
	if jiraErr != nil {
		logger.Fatal(jiraErr)
	}

	tgClient, tgErr := telegram.NewTelegramClient()
	if tgErr != nil {
		logger.Errorf("failed to configure telegram client: %w", tgErr)
	}

	dispatcher := dispath.NewDispatcher(
		manager,
		handlers.NewJiraCommentIssue(jiraClient, engine),
		handlers.NewTelegramNotify(tgClient, engine),
		handlers.NewUtilsLog(),
	)

	srv := server.NewHttpServer(dispatcher)
	logger.Fatal(srv.Run())
}
