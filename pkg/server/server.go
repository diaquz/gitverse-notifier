package server

import (
	"context"
	"io"
	"net"
	"net/http"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events/parsers"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/pipeline"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	addr           string
	pipeline       *pipeline.Pipeline
	accounts       gin.Accounts
	logAllRequests bool
}

func NewHttpServer(pipeline *pipeline.Pipeline) *HttpServer {
	addr := net.JoinHostPort(config.GlobalConfig.BindHost, config.GlobalConfig.HTTPPort)

	return &HttpServer{
		addr:           addr,
		pipeline:       pipeline,
		logAllRequests: config.GlobalConfig.LogRequests,
		accounts: gin.Accounts{
			config.GlobalConfig.BasicAuthUser: config.GlobalConfig.BasicAuthPassword,
		},
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
		group.Use(gin.BasicAuth(s.accounts))
		group.GET("/health", s.health)
		group.POST("/event", s.handleEvent)
	}
}

func (s *HttpServer) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *HttpServer) handleEvent(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	body, ok := readRequestBody(ctx, ginCtx)
	if !ok {
		ginCtx.Status(http.StatusBadRequest)
		return
	}

	eventName := ginCtx.Request.Header.Get("X-Gitverse-Event")
	eventTypeName := ginCtx.Request.Header.Get("X-Gitverse-Event-Type")
	requestId := ginCtx.GetHeader("X-Gitverse-Delivery")

	if requestId != "" {
		ctx = logger.With(ctx, "request-id", requestId)
	}

	event, err := parsers.ParseEvent(ctx, eventName, eventTypeName, requestId, body)
	if err != nil {
		logger.Error(ctx, "failed to parse gitverse event", 
			"event-name", eventName, "event-type", eventTypeName,
			"err", err, "headers", ginCtx.Request.Header, "body", string(body))

		ginCtx.Status(http.StatusBadRequest)
		return
	}

	if s.logAllRequests {
		logger.Info(ctx, "gitverse event received", "event", event.Type, "body", string(body))
	}

	if err := s.pipeline.Enqueue(ctx, &event); err != nil {
		logger.Error(ctx, "failed to enqueue event",
			"err", err, "event", event.Type, "repository", event.Repository)
		ginCtx.Status(http.StatusInternalServerError)
		return
	}

	ginCtx.Status(http.StatusOK)
}

func readRequestBody(ctx context.Context, ginCtx *gin.Context) (body []byte, ok bool) {
	body, err := io.ReadAll(ginCtx.Request.Body)
	if err != nil {
		logger.Error(ctx, "failed to read request body", "headers", ginCtx.Request.Header, "err", err)
		return
	}

	return body, true
}
