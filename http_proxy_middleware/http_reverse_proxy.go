package http_proxy_middleware

import (
	"gatewat_web/dao"
	"gatewat_web/middleware"
	"gatewat_web/reverse_proxy"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// 匹配接入方式 基于请求信息
func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}

		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		lb, err := dao.LoadBalanceHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		trans, err := dao.TransportorHandler.GetTrans(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}

		//创建reverseproxy
		//使用 reverseproxy.ServceHttp(c.Request, c.Response)
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(c, lb, trans)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
		return
	}
}
