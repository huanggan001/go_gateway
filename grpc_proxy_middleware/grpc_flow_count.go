package grpc_proxy_middleware

import (
	"gatewat_web/dao"
	"gatewat_web/public"
	"google.golang.org/grpc"
	"log"
)

func GrpcFlowCountMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		//统计项 1 全站 2 服务 3 租户
		totalCounter, err := public.FlowCountHandler.GetCounter(public.FlowTotal)
		if err != nil {
			return err
		}
		totalCounter.Increase()

		serviceCounter, err := public.FlowCountHandler.GetCounter(public.FlowServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			return err
		}
		serviceCounter.Increase()

		if err := handler(srv, ss); err != nil {
			log.Printf("GrpcFlowCountMiddleware rpc failed with error %v\n", err)
			return err
		}
		return nil
	}
}
