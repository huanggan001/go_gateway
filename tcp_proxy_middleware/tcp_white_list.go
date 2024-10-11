package tcp_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/public"
	"strings"
)

func TCPWhiteListMiddleware() func(c *TcpSliceRouterContext) {
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
		iplist := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			iplist = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(iplist) > 0 {
			if !public.InstringSlice(iplist, clientIp) {
				c.conn.Write([]byte(fmt.Sprintf("%s not in white ip list", clientIp)))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
