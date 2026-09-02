package server

import (
	"bytes"
	"io"
	"net"
	"net/http"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/logger"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	addr string
}

func NewHttpServer() *HttpServer {
	addr := net.JoinHostPort(config.GlobalConfig.BindHost, config.GlobalConfig.HTTPPort)

	return &HttpServer{
		addr: addr,
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
		logger.Errorf("failed to read /gitverse/event body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	logger.Infof("POST /gitverse/event headers=%v body=%s", ctx.Request.Header, string(body))

	// parse payload to gitverse event
	// call manager
	//   -> look up rules for event
	//   -> if jira enabled
	//   		parse task code, resolve task url
	//   		add comment to jira task
	//   		if needed make requesat to gitverse, to update title
	//   -> if telegram enabled
	//   		format massage (set link to task, if jira enabled and we now task number)
	//   		make request to bot
	// it all (after parsing) may be done in gorutin
	// -> 200
	ctx.Status(http.StatusOK)
}

