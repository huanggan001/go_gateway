package tcp_proxy_router

import (
	"context"
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/reverse_proxy"
	"gatewat_web/tcp_proxy_middleware"
	tcp_proxy "gatewat_web/tcp_server"
	"log"
	"net"
)

var tcpServerList = []*tcp_proxy.TcpServer{}

type tcpHandler struct {
}

func (t *tcpHandler) ServeTCP(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler----->\n"))
}

func TcpServerRun() {
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.TCPRule.Port)
			rb, err := dao.LoadBalanceHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf("[INFO] GetTcpLoadBalancer %v err: %v\n", addr, err)
				return
			}

			//构建路由及设置中间件
			router := tcp_proxy_middleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcp_proxy_middleware.TCPFlowCountMiddleware(),
				tcp_proxy_middleware.TCPFlowLimiterMiddleware(),
				tcp_proxy_middleware.TCPWhiteListMiddleware(),
				tcp_proxy_middleware.TCPBlackListMiddleware(),
			)

			//构建回调handler
			routerHandler := tcp_proxy_middleware.NewTcpSliceRouterHandler(func(c *tcp_proxy_middleware.TcpSliceRouterContext) tcp_proxy.TCPHandler {
				return reverse_proxy.NewTcpLoadBalanceReversProxy(c, rb)
			}, router)

			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)
			tcpServer := &tcp_proxy.TcpServer{
				Addr:    addr,
				Handler: routerHandler,
				BaseCtx: baseCtx,
			}
			tcpServerList = append(tcpServerList, tcpServer)
			log.Printf("[INFO] tcp_proxy_run %v\n", addr)
			if err := tcpServer.ListenAndServe(); err != nil && err != tcp_proxy.ErrServerClosed {
				log.Fatalf("[INFO] tcp_proxy_run %v err : %v\n", addr, err)
			}
		}(tempItem)
	}
}

func TcpServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf("[INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
	}
}
