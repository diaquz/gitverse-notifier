package main

import (
	"flag"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/events/actions"
	"gitverse-notifier/pkg/integrations/jira"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/server"
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

	var jiraClient *jira.JiraClient
	if client, err := jira.NewJiraClient(); err != nil {
		logger.Warnf("jira client not configured: %v", err)
	} else {
		jiraClient = client
	}

	dispatcher := events.NewDispatcher(
		manager,
		actions.NewJiraCommentIssue(jiraClient),
	)

	srv := server.NewHttpServer(dispatcher)
	logger.Fatal(srv.Run())
}
