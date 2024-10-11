package grpc_proxy_middleware

import (
	"encoding/json"
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/public"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
)

// jwt auth token
func GrpcJwtFlowCountMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errors.New("miss metadata from context")
		}

		appInfos := md.Get("app")
		if len(appInfos) == 0 {
			if err := handler(srv, ss); err != nil {
				log.Printf("GrpcJwtFlowCountMiddleware len(appInfos) == 0 rpc failed with error %v\n", err)
				return err
			}
			return nil
		}
		appInfo := &dao.App{}
		if err := json.Unmarshal([]byte(appInfos[0]), appInfo); err != nil {
			return err
		}

		appCounter, err := public.FlowCountHandler.GetCounter(public.FlowAppPrefix + appInfo.AppID)
		if err != nil {
			return err
		}
		appCounter.Increase()

		if appInfo.Qpd > 0 && appCounter.TotalCount > appInfo.Qpd {
			return errors.New(fmt.Sprintf("租户日请求量限流 limit:%v current:%v", appInfo.Qpd, appCounter.TotalCount))
		}

		if err := handler(srv, ss); err != nil {
			log.Printf("GrpcJwtFlowCountMiddleware rpc failed with error %v\n", err)
			return err
		}
		return nil
	}
}
