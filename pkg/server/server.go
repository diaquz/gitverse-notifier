package server

import (
	"context"
	"fmt"
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

const maxEventBody = 1 << 20

type HttpServer struct {
	addr           string
	pipeline       *pipeline.Pipeline
	accounts       gin.Accounts
	logAllRequests bool
	server         *http.Server
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
	if s.server != nil {
		return fmt.Errorf("server already running")
	}

	eng := s.registerRouter()
	s.server = &http.Server{Addr: s.addr, Handler: eng}
	defer s.clearServer()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return fmt.Errorf("no running http server")
	}

	return s.server.Shutdown(ctx)
}

func (s *HttpServer) clearServer() {
	s.server = nil
}

func (s *HttpServer) registerRouter() *gin.Engine {
	if config.GlobalConfig.LogLevel != "DEBUG" {
		gin.SetMode(gin.ReleaseMode)
	}

	eng := gin.New()
	eng.Use(gin.Recovery())
	eng.Use(gin.Logger())

	s.registerUtilRoutes(eng)
	s.registerGitverseRouterGroup(eng)
	s.registerPprof(eng)

	return eng
}

func (s *HttpServer) registerUtilRoutes(eng *gin.Engine) {
	group := eng.Group("/")

	{
		group.GET("/health", s.health)
	}
}

func (s *HttpServer) registerGitverseRouterGroup(eng *gin.Engine) {
	group := eng.Group("/gitverse")

	{
		group.Use(gin.BasicAuth(s.accounts))
		group.POST("/event", s.handleEvent)
	}
}

func (s *HttpServer) registerPprof(eng *gin.Engine) {
	if !config.GlobalConfig.EnablePprof {
		return
	}

	logger.Warn(nil, "pprof endpoints enabled", "path", "/debug/pprof/", "action", "server.setup")

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
	ginCtx.Request.Body = http.MaxBytesReader(ginCtx.Writer, ginCtx.Request.Body, maxEventBody)
	body, err := io.ReadAll(ginCtx.Request.Body)
	if err != nil {
		logger.Error(ctx, "failed to read request body", "headers", ginCtx.Request.Header, "err", err)
		return
	}

	return body, true
}
