package reverse_proxy

import (
	"gatewat_web/middleware"
	"gatewat_web/reverse_proxy/load_balance"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewLoadBalanceReverseProxy(c *gin.Context, lb load_balance.LoadBalance, trans *http.Transport) *httputil.ReverseProxy {

	//请求协调者
	director := func(req *http.Request) {
		nextAddr, err := lb.Get(req.URL.String())
		if err != nil {
			log.Fatal("get next addr fail: ", err)
		}

		target, err := url.Parse(nextAddr)
		if err != nil {
			log.Fatal(err)
		}
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		req.Host = target.Host
		//todo 当对域名（非内网）反向代理时需要设置此项，当作后端反向代理时不需要
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}
	//更改内容
	modifyFunc := func(resp *http.Response) error {

		//兼容websocket
		if strings.Contains(resp.Header.Get("Connection"), "Upgrade") {
			return nil
		}
		//var payload []byte
		//var readErr error
		//
		////兼容gzip压缩
		//if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		//	gr, err := gzip.NewReader(resp.Body)
		//	if err != nil {
		//		return err
		//	}
		//	payload, readErr = ioutil.ReadAll(gr)
		//	resp.Header.Del("Content-Encoding")
		//} else {
		//	payload, readErr = ioutil.ReadAll(resp.Body)
		//}
		//if readErr != nil {
		//	return readErr
		//}
		//
		////异常请求时设置StatusCode
		//if resp.StatusCode != 200 {
		//	payload = []byte("StatusCode error :" + string(payload))
		//}
		//
		////因为预读了数据所以内容重新回写
		//c.Set("status_code", resp.StatusCode)
		//c.Set("payload", payload)
		//resp.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		//resp.ContentLength = int64(len(payload))
		//resp.Header.Set("Content-Length", strconv.FormatInt(int64(len(payload)), 10))
		return nil
	}
	//错误回调：关闭real_server时测试， 错误回调
	//范围:transport.RoundTrip发生的错误、以及ModifyResponse发生的错误
	errFunc := func(w http.ResponseWriter, r *http.Request, err error) {
		middleware.ResponseError(c, 999, err)
	}
	return &httputil.ReverseProxy{Director: director, Transport: trans, ModifyResponse: modifyFunc, ErrorHandler: errFunc}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/") //该字符串是否以/结束
	bslash := strings.HasPrefix(b, "/") //该字符串是否以/开始
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
