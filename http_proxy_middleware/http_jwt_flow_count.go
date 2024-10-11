package http_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPJwtFlowCounterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appInterface, ok := c.Get("appInfo")
		if !ok {
			c.Next()
			return
		}

		appInfo := appInterface.(*dao.App)
		appCounter, err := public.FlowCountHandler.GetCounter(public.FlowAppPrefix + appInfo.AppID)
		if err != nil {
			middleware.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		appCounter.Increase()

		if appInfo.Qpd > 0 && appCounter.TotalCount > appInfo.Qpd {
			middleware.ResponseError(c, 2003, errors.New(fmt.Sprintf("租户日请求量限流 limit:%v current:%v", appInfo.Qpd, appCounter.TotalCount)))
			c.Abort()
			return
		}

		c.Next()
	}
}
