package tcp_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/public"
	"strings"
)

func TCPBlackListMiddleware() func(c *TcpSliceRouterContext) {
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

		whiteIpList := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			whiteIpList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		blackIpList := []string{}
		if serviceDetail.AccessControl.BlackList != "" {
			blackIpList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(blackIpList) > 0 {
			if !public.InstringSlice(whiteIpList, clientIp) && public.InstringSlice(blackIpList, clientIp) {
				c.conn.Write([]byte(fmt.Sprintf("%s in black ip list", clientIp)))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
