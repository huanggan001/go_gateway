package reverse_proxy

import (
	"context"
	"gatewat_web/reverse_proxy/load_balance"
	"gatewat_web/tcp_proxy_middleware"
	"io"
	"log"
	"net"
	"time"
)

func NewTcpLoadBalanceReversProxy(c *tcp_proxy_middleware.TcpSliceRouterContext, lb load_balance.LoadBalance) *TcpReverseProxy {
	return func() *TcpReverseProxy {
		nextAddr, err := lb.Get("")
		if err != nil {
			log.Fatal("get next addr fail")
		}
		return &TcpReverseProxy{
			ctx:             c.Ctx,
			Addr:            nextAddr,
			KeepAlivePeriod: time.Second,
			DialTimeout:     time.Second,
		}
	}() //通过在匿名函数后面添加 ()，就可以立即执行这个匿名函数，并将执行结果作为整个表达式的值返回。
}

type TcpReverseProxy struct {
	ctx                  context.Context //单次请求单独设置
	Addr                 string
	KeepAlivePeriod      time.Duration //设置
	DialTimeout          time.Duration //设置超时时间
	DialContext          func(ctx context.Context, network, address string) (net.Conn, error)
	OnDialError          func(src net.Conn, dstDialErr error)
	ProxyProtocolVersion int
}

func (dp *TcpReverseProxy) dialTimeout() time.Duration {
	if dp.DialTimeout > 0 {
		return dp.DialTimeout
	}
	return 10 * time.Second
}

var defaultDialer = new(net.Dialer)

func (dp *TcpReverseProxy) dialContext() func(ctx context.Context, network, address string) (net.Conn, error) {
	if dp.DialContext != nil {
		return dp.DialContext
	}
	return (&net.Dialer{
		Timeout:   dp.DialTimeout,     //连接超时
		KeepAlive: dp.KeepAlivePeriod, //设置连接的检测时长
	}).DialContext
}

func (dp *TcpReverseProxy) keepAlivePeriod() time.Duration {
	if dp.KeepAlivePeriod > 0 {
		return dp.KeepAlivePeriod
	}
	return time.Minute
}

// 传入上游conn, 在这里完成下游连接与数据交换
func (dp *TcpReverseProxy) ServeTCP(ctx context.Context, src net.Conn) {
	//设置连接超时
	var cancel context.CancelFunc
	if dp.DialTimeout >= 0 {
		ctx, cancel = context.WithTimeout(ctx, dp.DialTimeout)
	}
	dst, err := dp.dialContext()(ctx, "tcp", dp.Addr) //连接下游服务器，代理服务器与目标服务器之间建立的连接
	if cancel != nil {
		cancel()
	}
	if err != nil {
		dp.onDialError()(src, err)
		return
	}
	defer func() { go dst.Close() }() // 记得退出下游服务

	//设置dst的keepAlive 参数， 在数据请求之前
	if ka := dp.keepAlivePeriod(); ka > 0 {
		if c, ok := dst.(*net.TCPConn); ok {
			c.SetKeepAlive(true)
			c.SetKeepAlivePeriod(ka)
		}
	}
	errc := make(chan error, 1)
	go dp.proxyCopy(errc, src, dst)
	go dp.proxyCopy(errc, dst, src)
	<-errc
}

func (dp *TcpReverseProxy) onDialError() func(src net.Conn, dstDialErr error) {
	if dp.OnDialError != nil {
		return dp.OnDialError
	}
	return func(src net.Conn, dstDialErr error) {
		log.Printf("tcpproxy: for incoming conn %v, error dialing %q:%v", src.RemoteAddr().String(), dp.Addr, dstDialErr)
		src.Close()
	}
}

func (dp *TcpReverseProxy) proxyCopy(errc chan<- error, dst, src net.Conn) {
	_, err := io.Copy(dst, src)
	errc <- err
}
