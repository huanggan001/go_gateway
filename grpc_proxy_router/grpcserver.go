package grpc_proxy_router

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/grpc_proxy_middleware"
	"gatewat_web/reverse_proxy"
	"github.com/e421083458/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"log"
	"net"
)

var grpcServerList = []*warpGrpcServer{}

// go get -u google.golang.org/grpc
type warpGrpcServer struct {
	Addr string
	*grpc.Server
}

func GrpcServerRun() {
	serviceList := dao.ServiceManagerHandler.GetGrpcServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.GRPCRule.Port)
			rb, err := dao.LoadBalanceHandler.GetLoadBalancer(serviceDetail)

			if err != nil {
				log.Fatalf("[INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatalf("[INFO] GrpcListen %v err:%v\n", addr, err)
			}
			grpcHandler := reverse_proxy.NewGrpcLoadBalanceHandler(rb)
			s := grpc.NewServer(
				grpc.ChainStreamInterceptor(
					grpc_proxy_middleware.GrpcFlowCountMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcFlowLimiterMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtAuthTokenMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtFlowCountMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtFlowLimiterMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcWhiteListMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcBlackListMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcHeaderTransferMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcReverseProxyMiddleware(rb)),
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler))

			grpcServerList = append(grpcServerList, &warpGrpcServer{
				Addr:   addr,
				Server: s,
			})
			log.Printf("[INFO] grpc_proxy_run %v\n", addr)
			if err := s.Serve(listener); err != nil {
				log.Fatalf("[INFO] grpc_proxy_run %v err :%v\n", addr, err)
			}

		}(tempItem)
	}
}

func GrpcServerStop() {
	for _, grpcServer := range grpcServerList {
		grpcServer.GracefulStop()
		log.Printf("[INFO] grpc_proxy_stop %v stopped", grpcServer.Addr)
	}
}
