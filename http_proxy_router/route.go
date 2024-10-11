package http_proxy_router

import (
	"gatewat_web/controller"
	"gatewat_web/http_proxy_middleware"
	"gatewat_web/middleware"
	"github.com/gin-gonic/gin"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {

	//router := gin.Default()
	router := gin.New()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	oauth := router.Group("/oauth")
	oauth.Use(middleware.TranslationMiddleware())
	{
		controller.OauthRegister(oauth)
	}

	router.Use(http_proxy_middleware.HTTPAccessModeMiddleware(),
		http_proxy_middleware.HTTPFlowCountMiddleware(),
		http_proxy_middleware.HTTPFlowLimiterMiddleware(),

		http_proxy_middleware.HTTPJwtAuthTokenMiddleware(),
		http_proxy_middleware.HTTPJwtFlowCounterMiddleware(),
		http_proxy_middleware.HTTPJwtFlowLimiterMiddleware(),
		http_proxy_middleware.HTTPBlackListMiddleware(),
		http_proxy_middleware.HTTPWhiteListMiddleware(),

		http_proxy_middleware.HTTPHeaderTransferMiddleware(),
		http_proxy_middleware.HTTPStripUrlMiddleware(),
		http_proxy_middleware.HTTPUrlRewriteMiddleware(),
		http_proxy_middleware.HTTPReverseProxyMiddleware())
	return router
}
