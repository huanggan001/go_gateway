package http_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPFlowLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := public.FlowLimiterHandler.GetLimiter(serviceDetail.Info.ServiceName, float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				middleware.ResponseError(c, 2002, err)
				return
			}
			if !serviceLimiter.Allow() {
				middleware.ResponseError(c, 2003, errors.New(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				c.Abort()
				return
			}
		}

		if serviceDetail.AccessControl.ClientIPFlowLimit != 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(serviceDetail.Info.ServiceName+"_"+c.ClientIP(), float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				middleware.ResponseError(c, 2004, err)
				return
			}
			if !clientLimiter.Allow() {
				middleware.ResponseError(c, 2005, errors.New(fmt.Sprintf("%v flow limit %v", c.ClientIP(), serviceDetail.AccessControl.ClientIPFlowLimit)))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
