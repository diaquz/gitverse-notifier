package main

import (
	"flag"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/events/actions"
	"gitverse-notifier/pkg/events/enrichers"
	"gitverse-notifier/pkg/events/parsers"
	"gitverse-notifier/pkg/integrations/jira"
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

	manager, err := events.SetupActionsManager()
	if err != nil {
		logger.Fatal(err)
	}

	repositoriesManager, err := repositories.SetupRepositoriesManager()
	if err != nil {
		logger.Fatal(err)
	}

	engine, err := templates.SetupTemplateEngine()
	if err != nil {
		logger.Fatal(err)
	}

	if err := parsers.SetupEventParser(
		enrichers.NewJiraIssueKeys(repositoriesManager),
		enrichers.NewGitverseLinks(repositoriesManager),
	); err != nil {
		logger.Fatal(err)
	}

	jiraClient, jiraErr := jira.NewJiraClient()
	if jiraErr != nil {
		logger.Fatal(jiraErr)
	}

	dispatcher := events.NewDispatcher(
		manager,
		actions.NewJiraCommentIssue(jiraClient, engine),
		actions.NewTelegramNotify(engine),
	)

	srv := server.NewHttpServer(dispatcher)
	logger.Fatal(srv.Run())
}
