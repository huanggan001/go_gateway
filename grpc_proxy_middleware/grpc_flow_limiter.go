package grpc_proxy_middleware

import (
	"gatewat_web/dao"
	"gatewat_web/public"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"strings"
)

func GrpcFlowLimiterMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := public.FlowLimiterHandler.GetLimiter(serviceDetail.Info.ServiceName, float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				return err
			}
			if !serviceLimiter.Allow() {
				return err
			}
		}
		peerCtx, ok := peer.FromContext(ss.Context())
		if !ok {
			return errors.New("peer not found with context")
		}
		peerAddr := peerCtx.Addr.String()
		addrPort := strings.LastIndex(peerAddr, ":")
		clientIP := peerAddr[0:addrPort]
		if serviceDetail.AccessControl.ClientIPFlowLimit != 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(serviceDetail.Info.ServiceName+"_"+clientIP, float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				return err
			}
			if !clientLimiter.Allow() {
				return err
			}
		}

		if err := handler(srv, ss); err != nil {
			log.Printf("GrpcFlowLimiterMiddleware rpc failed with error %v\n", err)
			return err
		}
		return nil
	}
}
