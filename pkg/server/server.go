package server

import (
	"context"
	"io"
	"net"
	"net/http"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/dispath"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/events/parsers"
	"gitverse-notifier/pkg/logger"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	addr           string
	processor      *dispath.Processor
	accounts       gin.Accounts
	logAllRequests bool
}

func NewHttpServer(proc *dispath.Processor) *HttpServer {
	addr := net.JoinHostPort(config.GlobalConfig.BindHost, config.GlobalConfig.HTTPPort)

	return &HttpServer{
		addr:           addr,
		processor: processor,
		logAllRequests: config.GlobalConfig.LogRequests,
		addr:      addr,
		processor: proc,
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
	ctx := ctxWithRequestId(ginCtx)
	body, ok := readRequestBody(ctx, ginCtx)
	if !ok {
		ginCtx.Status(http.StatusBadRequest)
		return
	}

	eventName := ginCtx.Request.Header.Get("X-Gitverse-Event")
	eventTypeName := ginCtx.Request.Header.Get("X-Gitverse-Event-Type")
	event, err := parsers.ParseEvent(ctx, eventName, eventTypeName, body)
	if err != nil {
		logger.Error(ctx, "failed to parse gitverse event", "err", err,
			"headers", ginCtx.Request.Header, "body", string(body))
		ginCtx.Status(http.StatusBadRequest)
		return
	}

	if s.logAllRequests {
		logger.Info(ctx, "gitverse event received",
			"headers", ginCtx.Request.Header, "body", string(body))
	}

	ginCtx.Status(http.StatusOK)

	go func(ctx context.Context, event events.Event) {
		s.processor.Process(ctx, event)
	}(ctx, event)
}

func ctxWithRequestId(ginCtx *gin.Context) (ctx context.Context) {
	ctx = context.WithoutCancel(ginCtx.Request.Context())
	if delivery := ginCtx.GetHeader("X-Gitverse-Delivery"); delivery != "" {
		ctx = logger.With(ctx, "request-id", delivery)
	}

	return
}

func readRequestBody(ctx context.Context, ginCtx *gin.Context) (body []byte, ok bool) {
	body, err := io.ReadAll(ginCtx.Request.Body)
	if err != nil {
		logger.Error(ctx, "failed to read request body", "err", err)
		return
	}

	return body, true
}
