package main

import (
	commonconfig "eco-knock-be-embedded/internal/common/config"
	"eco-knock-be-embedded/internal/common/middleware"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startRESTServer(commonConfig commonconfig.CommonConfig) error {
	router := newRESTRouter()
	return router.Run(net.JoinHostPort("", formatPort(commonConfig.ServerHTTPPort)))
}

func newRESTRouter() *gin.Engine {
	router := gin.Default()
	router.Use(middleware.HandleErrors())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}
