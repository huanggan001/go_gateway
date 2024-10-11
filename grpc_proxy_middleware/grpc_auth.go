package grpc_proxy_middleware

import (
	"google.golang.org/grpc"
	"log"
)

// 流式拦截器
func GrpcAuthStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	err := handler(srv, ss)
	if err != nil {
		log.Printf("rpc failed with error %v\n", err)
	}
	return err
}
