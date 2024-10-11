package main

import (
	"flag"
	"fmt"
	"gatewat_web/common/lib"
	"gatewat_web/dao"
	"gatewat_web/grpc_proxy_router"
	"gatewat_web/http_proxy_router"
	"gatewat_web/router"
	"gatewat_web/tcp_proxy_router"
	"os"
	"os/signal"
	"syscall"
)

//endpoint dashboard后台管理 server代理服务器
//config ./conf/dev/ 对应配置文件夹

// go run main.go -config="./conf/dev/" -endpoint="dashboard"
// go run main.go -config="./conf/dev/" -endpoint="server"
var (
	endpoint = flag.String("endpoint", "", "input dashboard or server")
	config   = flag.String("config", "", "input config file like ./conf/dev/")
)

func main() {
	flag.Parse()

	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *endpoint == "dashboard" {
		lib.InitModule(*config)
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		router.HttpServerStop()
	} else {
		lib.InitModule(*config)
		defer lib.Destroy()
		fmt.Println("start server")
		dao.ServiceManagerHandler.LoadOnce()
		dao.AppManagerHandler.LoadOnce()
		go func() {
			http_proxy_router.HttpServerRun()
		}()
		go func() {
			http_proxy_router.HttpsServerRun()
		}()
		go func() {
			tcp_proxy_router.TcpServerRun()
		}()
		go func() {
			grpc_proxy_router.GrpcServerRun()
		}()
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		tcp_proxy_router.TcpServerStop()
		http_proxy_router.HttpServerStop()
		http_proxy_router.HttpsServerStop()
		grpc_proxy_router.GrpcServerStop()

	}
}
