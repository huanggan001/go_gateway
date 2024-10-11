package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	dlog "github.com/e421083458/golang_common/log"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var TimeLocation *time.Location
var TimeFormat = "2006-01-02 15:04:05"
var DateFormat = "2006-01-02"
var LocalIP = net.ParseIP("127.0.0.1")

// 公共初始化函数：支持两种方式设置配置文件
//
// 函数传入配置文件 Init("./conf/dev/")
// 如果配置文件为空，会从命令行中读取 	  -config conf/dev/
func InitModule(configPath string) error {
	return initModule(configPath, []string{"base", "mysql", "redis"})
}

// 模块初始化
func initModule(configPath string, modules []string) error {
	//conf := flag.String("config", configPath, "input config file like ./conf/dev/")
	//flag.Parse()
	//if *conf == "" {
	//	flag.Usage()
	//	os.Exit(1)
	//}
	if configPath == "" {
		fmt.Printf("input config file like ./conf/dev/")
		os.Exit(1)
	}

	log.Println("------------------------------------------------------------------------")
	log.Printf("[INFO]  config=%s\n", configPath)
	log.Printf("[INFO] %s\n", " start loading resources.")

	// 设置ip信息，优先设置便于日志打印
	ips := GetLocalIPs()
	if len(ips) > 0 {
		LocalIP = ips[0]
	}

	// 解析配置文件目录
	if err := ParseConfPath(configPath); err != nil {
		return err
	}

	//初始化配置文件
	if err := InitViperConf(); err != nil {
		return err
	}

	// 加载base配置
	if InArrayString("base", modules) {
		if err := InitBaseConf(GetConfPath("base")); err != nil {
			fmt.Printf("[ERROR] %s%s\n", time.Now().Format(TimeFormat), " InitBaseConf:"+err.Error())
		}
	}

	// 加载redis配置
	if InArrayString("redis", modules) {
		if err := InitRedisConf(GetConfPath("redis_map")); err != nil {
			fmt.Printf("[ERROR] %s%s\n", time.Now().Format(TimeFormat), " InitRedisConf:"+err.Error())
		}
	}

	// 加载mysql配置并初始化实例
	if InArrayString("mysql", modules) {
		if err := InitDBPool(GetConfPath("mysql_map")); err != nil {
			fmt.Printf("[ERROR] %s%s\n", time.Now().Format(TimeFormat), " InitDBPool:"+err.Error())
		}
	}

	// 设置时区
	if location, err := time.LoadLocation(ConfBase.TimeLocation); err != nil {
		return err
	} else {
		TimeLocation = location
	}

	log.Printf("[INFO] %s\n", " success loading resources.")
	log.Println("------------------------------------------------------------------------")
	return nil
}

// 公共销毁函数
func Destroy() {
	log.Println("------------------------------------------------------------------------")
	log.Printf("[INFO] %s\n", " start destroy resources.")
	CloseDB()
	dlog.Close()
	log.Printf("[INFO] %s\n", " success destroy resources.")
}

func HttpGET(trace *TraceContext, urlString string, urlParams url.Values, msTimeout int, header http.Header) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{
		Timeout: time.Duration(msTimeout) * time.Millisecond,
	}
	urlString = AddGetDataToUrl(urlString, urlParams)
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	if len(header) > 0 {
		req.Header = header
	}
	req = addTrace2Header(req, trace)
	resp, err := client.Do(req)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"result":    Substr(string(body), 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	Log.TagInfo(trace, DLTagHTTPSuccess, map[string]interface{}{
		"url":       urlString,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "GET",
		"args":      urlParams,
		"result":    Substr(string(body), 0, 1024),
	})
	return resp, body, nil
}

func HttpPOST(trace *TraceContext, urlString string, urlParams url.Values, msTimeout int, header http.Header, contextType string) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{
		Timeout: time.Duration(msTimeout) * time.Millisecond,
	}
	if contextType == "" {
		contextType = "application/x-www-form-urlencoded"
	}
	urlParamEncode := urlParams.Encode()
	req, err := http.NewRequest("POST", urlString, strings.NewReader(urlParamEncode))
	if len(header) > 0 {
		req.Header = header
	}
	req = addTrace2Header(req, trace)
	req.Header.Set("Content-Type", contextType)
	resp, err := client.Do(req)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(urlParamEncode, 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(urlParamEncode, 0, 1024),
			"result":    Substr(string(body), 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	Log.TagInfo(trace, DLTagHTTPSuccess, map[string]interface{}{
		"url":       urlString,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "POST",
		"args":      Substr(urlParamEncode, 0, 1024),
		"result":    Substr(string(body), 0, 1024),
	})
	return resp, body, nil
}

func HttpJSON(trace *TraceContext, urlString string, jsonContent string, msTimeout int, header http.Header) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{
		Timeout: time.Duration(msTimeout) * time.Millisecond,
	}
	req, err := http.NewRequest("POST", urlString, strings.NewReader(jsonContent))
	if len(header) > 0 {
		req.Header = header
	}
	req = addTrace2Header(req, trace)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(jsonContent, 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(jsonContent, 0, 1024),
			"result":    Substr(string(body), 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	Log.TagInfo(trace, DLTagHTTPSuccess, map[string]interface{}{
		"url":       urlString,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "POST",
		"args":      Substr(jsonContent, 0, 1024),
		"result":    Substr(string(body), 0, 1024),
	})
	return resp, body, nil
}

func AddGetDataToUrl(urlString string, data url.Values) string {
	if strings.Contains(urlString, "?") {
		urlString = urlString + "&"
	} else {
		urlString = urlString + "?"
	}
	return fmt.Sprintf("%s%s", urlString, data.Encode())
}

func addTrace2Header(request *http.Request, trace *TraceContext) *http.Request {
	traceId := trace.TraceId
	cSpanId := NewSpanId()
	if traceId != "" {
		request.Header.Set("didi-header-rid", traceId)
	}
	if cSpanId != "" {
		request.Header.Set("didi-header-spanid", cSpanId)
	}
	trace.CSpanId = cSpanId
	return request
}

func GetMd5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Encode(data string) (string, error) {
	h := md5.New()
	_, err := h.Write([]byte(data))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ParseServerAddr(serverAddr string) (host, port string) {
	serverInfo := strings.Split(serverAddr, ":")
	if len(serverInfo) == 2 {
		host = serverInfo[0]
		port = serverInfo[1]
	} else {
		host = serverAddr
		port = ""
	}
	return host, port
}

// NewTrace 创建新的 TraceContext 对象，并生成唯一的 TraceId 和 SpanId
// 分布式链路追踪：通过生成唯一的 TraceId 和 SpanId，
// 可以在监控系统或日志系统中跟踪每个请求在整个系统中的路径，识别请求是如何流经各个服务的。
func NewTrace() *TraceContext {
	trace := &TraceContext{}
	trace.TraceId = GetTraceId() //唯一标识
	trace.SpanId = NewSpanId()
	return trace
}

// NewSpanId 生成唯一的 SpanId。
// 通过 XOR 操作结合 IP 地址的整数值和当前时间戳，再加上一个随机数，构造出一个唯一的 SpanId。
// 这可以帮助精细化追踪每个请求操作中的步骤，特别是在分布式环境中非常有用。
func NewSpanId() string {
	timestamp := uint32(time.Now().Unix())
	ipToLong := binary.BigEndian.Uint32(LocalIP.To4())
	b := bytes.Buffer{}
	b.WriteString(fmt.Sprintf("%08x", ipToLong^timestamp))
	b.WriteString(fmt.Sprintf("%08x", rand.Int31()))
	return b.String()
}

// GetTraceId 基于本地 IP 地址、时间戳等信息生成的，确保每个请求有唯一的标识。
func GetTraceId() (traceId string) {
	return calcTraceId(LocalIP.String())
}

// calcTraceId 函数中，结合了时间戳、进程 ID (pid) 以及随机数，再加上 IP 地址，
// 生成了一个具有高唯一性的追踪 ID，确保同一台机器不同请求的追踪 ID 也是唯一的。
func calcTraceId(ip string) (traceId string) {
	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getpid()

	b := bytes.Buffer{}
	netIP := net.ParseIP(ip)
	if netIP == nil {
		b.WriteString("00000000")
	} else {
		b.WriteString(hex.EncodeToString(netIP.To4()))
	}
	b.WriteString(fmt.Sprintf("%08x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%06x", rand.Int31n(1<<24)))
	b.WriteString("b0") // 末两位标记来源,b0为go

	return b.String()
}

func GetLocalIPs() (ips []net.IP) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}
	return ips
}

func InArrayString(s string, arr []string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}

// Substr 字符串的截取
func Substr(str string, start int64, end int64) string {
	length := int64(len(str))
	if start < 0 || start > length {
		return ""
	}
	if end < 0 {
		return ""
	}
	if end > length {
		end = length
	}
	return string(str[start:end])
}
