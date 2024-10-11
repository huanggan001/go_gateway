package grpc_proxy_middleware

import (
	"context"
	"gatewat_web/reverse_proxy/load_balance"
	"github.com/e421083458/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"log"
)

func GrpcReverseProxyMiddleware(lb load_balance.LoadBalance) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 获取下一个地址
		nextAddr, err := lb.Get("")
		if err != nil {
			log.Fatalf("Failed to get next address: %v", err)
		}

		// 创建 director 函数
		director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
			// 使用下一个地址建立 gRPC 连接
			c, err := grpc.DialContext(ctx, nextAddr, grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
			return ctx, c, err
		}

		// 创建 TransparentHandler
		transparentHandler := proxy.TransparentHandler(director)

		// 调用下一个拦截器或处理程序
		return transparentHandler(srv, ss)
	}
}
