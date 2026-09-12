package server

import (
	"context"
	"io"
	"net"
	"net/http"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/events/parsers"
	"gitverse-notifier/pkg/logger"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	addr       string
	dispatcher *events.Dispatcher
}

func NewHttpServer(dispatcher *events.Dispatcher) *HttpServer {
	addr := net.JoinHostPort(config.GlobalConfig.BindHost, config.GlobalConfig.HTTPPort)

	return &HttpServer{
		addr:       addr,
		dispatcher: dispatcher,
	}
}

func (s *HttpServer) Run() error {
	eng := s.registerRouter()

	return http.ListenAndServe(s.addr, eng)
}

func (s *HttpServer) registerRouter() *gin.Engine {
	if config.GlobalConfig.LogLevel != "DEBUG" {
		gin.SetMode(gin.ReleaseMode)
	}

	eng := gin.New()
	eng.Use(gin.Recovery())
	eng.Use(gin.Logger())

	s.registerGitverseRouterGroup(eng)

	return eng
}

func (s *HttpServer) registerGitverseRouterGroup(eng *gin.Engine) {
	group := eng.Group("/gitverse")

	{
		group.GET("/health", s.health)
		group.POST("/event", s.handleEvent)
	}
}

func (s *HttpServer) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *HttpServer) handleEvent(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.Errorf("failed to read request body: %v", err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	eventName := ctx.Request.Header.Get("X-Gitverse-Event")
	eventTypeName := ctx.Request.Header.Get("X-Gitverse-Event-Type")
	event, err := parsers.ParseEvent(eventName, eventTypeName, body)
	if err != nil {
		logger.Errorf("failed to parse gitverse event: %v", err)
		logger.Errorf("headers=%v body=%s", ctx.Request.Header, string(body))
	}

	ctx.Status(http.StatusOK)

	go func(event events.Event) {
		s.dispatcher.Dispatch(context.Background(), event)
	}(event)
}
