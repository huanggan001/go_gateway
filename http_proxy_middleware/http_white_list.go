package http_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

func HTTPWhiteListMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		iplist := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			iplist = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(iplist) > 0 {
			if !public.InstringSlice(iplist, c.ClientIP()) {
				middleware.ResponseError(c, 3002, errors.New(fmt.Sprintf("%s not in white ip list", c.ClientIP())))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
