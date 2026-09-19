package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/pprof"

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
	s.registerPprof(eng)

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

func (s *HttpServer) registerPprof(eng *gin.Engine) {
	if !config.GlobalConfig.EnablePprof {
		return
	}

	logger.Warn(nil, "pprof endpoints enabled",
		"path", "/debug/pprof/",
		"action", "server_setup")

	group := eng.Group("/debug/pprof", gin.BasicAuth(s.accounts))
	{
		group.GET("/", gin.WrapF(pprof.Index))
		group.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		group.GET("/profile", gin.WrapF(pprof.Profile))
		group.GET("/symbol", gin.WrapF(pprof.Symbol))
		group.POST("/symbol", gin.WrapF(pprof.Symbol))
		group.GET("/trace", gin.WrapF(pprof.Trace))
		group.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		group.GET("/block", gin.WrapH(pprof.Handler("block")))
		group.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		group.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		group.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		group.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
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
