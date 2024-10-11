package tcp_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/public"
	"github.com/pkg/errors"
	"strings"
)

func TCPFlowLimiterMiddleware() func(c *TcpSliceRouterContext) {
	return func(c *TcpSliceRouterContext) {
		serviceInterface := c.Get("service")
		if serviceInterface == nil {
			c.conn.Write([]byte("get service empty"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		splits := strings.Split(c.conn.RemoteAddr().String(), ":")
		clientIp := ""
		if len(splits) == 2 {
			clientIp = splits[0]
		}
		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowServicePrefix+serviceDetail.Info.ServiceName, float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				c.conn.Write([]byte(err.Error()))
				return
			}
			if !serviceLimiter.Allow() {
				c.conn.Write([]byte(errors.New(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit)).Error()))
				c.Abort()
				return
			}
		}

		if serviceDetail.AccessControl.ClientIPFlowLimit != 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowServicePrefix+serviceDetail.Info.ServiceName+"_"+clientIp, float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				c.conn.Write([]byte(err.Error()))
				return
			}
			if !clientLimiter.Allow() {
				c.conn.Write([]byte(errors.New(fmt.Sprintf("%v flow limit %v", clientIp, serviceDetail.AccessControl.ClientIPFlowLimit)).Error()))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
