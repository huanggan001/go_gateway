package grpc_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/public"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"strings"
)

func GrpcBlackListMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		whiteIpList := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			whiteIpList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		blackIpList := []string{}
		if serviceDetail.AccessControl.BlackList != "" {
			blackIpList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}

		peerCtx, ok := peer.FromContext(ss.Context())
		if !ok {
			return errors.New("peer not found with context")
		}
		peerAddr := peerCtx.Addr.String()
		addrPos := strings.LastIndex(peerAddr, ":")
		clientIP := peerAddr[0:addrPos]

		if serviceDetail.AccessControl.OpenAuth == 1 && len(blackIpList) > 0 {
			if !public.InstringSlice(whiteIpList, clientIP) && public.InstringSlice(blackIpList, clientIP) {

				return errors.New(fmt.Sprintf("%s in black ip list", clientIP))
			}
		}

		if err := handler(srv, ss); err != nil {
			log.Printf("GrpcBlackListMiddleware rpc failed with error %v\n", err)
			return err
		}
		return nil
	}
}
