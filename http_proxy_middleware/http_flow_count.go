package http_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

func HTTPFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//统计项 1 全站 2 服务 3 租户
		totalCounter, err := public.FlowCountHandler.GetCounter(public.FlowTotal)
		if err != nil {
			middleware.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		totalCounter.Increase()

		dayCount, _ := totalCounter.GetDayData(time.Now())
		fmt.Printf("totalCounter qps :%v, dayCount:%v\n", totalCounter.QPS, dayCount)

		serviceCounter, err := public.FlowCountHandler.GetCounter(public.FlowServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		serviceCounter.Increase()

		ServiceCount, _ := serviceCounter.GetDayData(time.Now())
		fmt.Printf("serviceCounter qps :%v, ServiceCount:%v ", serviceCounter.QPS, ServiceCount)

		c.Next()
	}
}
