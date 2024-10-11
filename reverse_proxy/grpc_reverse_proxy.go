package reverse_proxy

import (
	"context"
	"fmt"
	"gatewat_web/reverse_proxy/load_balance"
	"github.com/e421083458/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"log"
)

func NewGrpcLoadBalanceHandler(lb load_balance.LoadBalance) grpc.StreamHandler {
	return func() grpc.StreamHandler {
		nextAddr, err := lb.Get("")
		//nextAddr := "127.0.0.1:50055"
		fmt.Println("grpc服务端地址：", nextAddr)
		if err != nil {
			log.Fatalf("get next addr fail")
		}
		director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
			c, err := grpc.DialContext(ctx, nextAddr, grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
			return ctx, c, err

		}
		return proxy.TransparentHandler(director)
	}()
}
