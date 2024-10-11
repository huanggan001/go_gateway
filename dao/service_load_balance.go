package dao

import (
	"fmt"
	"gatewat_web/public"
	"gatewat_web/reverse_proxy/load_balance"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LoadBalance struct {
	ID            int64  `json:"id" gorm:"primary_key"`
	ServiceID     int64  `json:"service_id" gorm:"service_id" description:"服务id"`
	CheckMethod   int    `json:"check_method" gorm:"check_method" description:"检查方法 tcpchk=检测端口是否握手成功"`
	CheckTimeout  int    `json:"check_timeout" gorm:"check_timeout" description:"check超时方法"`
	CheckInterval int    `json:"check_interval" gorm:"check_interval" description:"检查间隔，单位s"`
	RoundType     int    `json:"round_type" gorm:"round_type" description:"轮询方式 round/weight_round/random/ip_hash"`
	IpList        string `json:"ip_list" gorm:"ip_list" description:"ip列表"`
	WeightList    string `json:"weight_list" gorm:"weight_list" description:"权重列表"`
	ForbidList    string `json:"forbid_list" gorm:"forbid_list" description:"禁用IP列表"`

	UpstreamConnectTimeout int `json:"upstream_connect_timeout" gorm:"upstream_connect_timeout" description:"下游建立连接超时，单位s"`
	UpstreamHeaderTimeout  int `json:"upstream_header_timeout" gorm:"upstream_header_timeout" description:"下游获取header超时，单位s"`
	UpstreamIdleTimeout    int `json:"upstream_idle_timeout" gorm:"upstream_idle_timeout" description:"下游链接最大空闲时间，单位s"`
	UpstreamMaxIdle        int `json:"upstream_max_idle" gorm:"upstream_max_idle" description:"下游最大空闲链接数"`
}

func (balance *LoadBalance) TableName() string {
	return "gateway_service_load_balance"
}

func (balance *LoadBalance) Find(c *gin.Context, db *gorm.DB, search *LoadBalance) (*LoadBalance, error) {
	model := &LoadBalance{}
	err := db.WithContext(c).Where(search).First(model).Error
	return model, err
}

func (balance *LoadBalance) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.WithContext(c).Save(balance).Error; err != nil {
		return err
	}
	return nil
}

func (balance *LoadBalance) GetIPListByModel() []string {
	return strings.Split(balance.IpList, ",")
}

func (t *LoadBalance) GetWeightListByModel() []string {
	return strings.Split(t.WeightList, ",")
}

type LoadBalancerItem struct {
	LoadBanlance load_balance.LoadBalance
	ServiceName  string
}

type LoadBalancer struct {
	LoadBalanceMap   map[string]*LoadBalancerItem
	LoadBalanceSlice []*LoadBalancerItem
	Locker           sync.RWMutex
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		LoadBalanceMap:   map[string]*LoadBalancerItem{},
		LoadBalanceSlice: []*LoadBalancerItem{},
		Locker:           sync.RWMutex{},
	}
}

var LoadBalanceHandler *LoadBalancer

func init() {
	LoadBalanceHandler = NewLoadBalancer()
}

func (lb *LoadBalancer) GetLoadBalancer(service *ServiceDetail) (load_balance.LoadBalance, error) {
	for _, lbItem := range lb.LoadBalanceSlice {
		if lbItem.ServiceName == service.Info.ServiceName {
			return lbItem.LoadBanlance, nil
		}
	}
	schema := "http://"
	if service.HTTPRule.NeedHttps == 1 {
		schema = "https://"
	}
	if service.Info.LoadType == public.LoadTypeTCP || service.Info.LoadType == public.LoadTypeGRPC {
		schema = ""
	}
	ipList := service.LoadBalance.GetIPListByModel()
	fmt.Println("ipList:", ipList)
	weightList := service.LoadBalance.GetWeightListByModel()
	fmt.Println("weightList:", weightList)
	ipConf := map[string]string{}
	for ipIndex, ipItem := range ipList {
		ipConf[ipItem] = weightList[ipIndex]
	}
	mConf, err := load_balance.NewLoadBalanceCheckConf(fmt.Sprintf("%s%s", schema, "%s"), ipConf)

	if err != nil {
		return nil, err
	}
	lb1 := load_balance.LoadBalanceFactorWithConf(load_balance.LbType(service.LoadBalance.RoundType), mConf)

	lbItem := &LoadBalancerItem{
		LoadBanlance: lb1,
		ServiceName:  service.Info.ServiceName,
	}
	lb.Locker.Lock()
	defer lb.Locker.Unlock()
	lb.LoadBalanceSlice = append(lb.LoadBalanceSlice, lbItem)
	lb.LoadBalanceMap[service.Info.ServiceName] = lbItem
	return lb1, nil
}

type Transportor struct {
	TransportMap   map[string]*TransportItem
	TransportSlice []*TransportItem
	Locker         sync.RWMutex
}

type TransportItem struct {
	Trans       *http.Transport
	ServiceName string
}

var TransportorHandler *Transportor

func NewTransportor() *Transportor {
	return &Transportor{
		TransportMap:   map[string]*TransportItem{},
		TransportSlice: []*TransportItem{},
		Locker:         sync.RWMutex{},
	}
}

func init() {
	TransportorHandler = NewTransportor()
}

func (t *Transportor) GetTrans(service *ServiceDetail) (*http.Transport, error) {
	for _, transItem := range t.TransportSlice {
		if transItem.ServiceName == service.Info.ServiceName {
			return transItem.Trans, nil
		}
	}
	//todo 优化点
	if service.LoadBalance.UpstreamConnectTimeout == 0 {
		service.LoadBalance.UpstreamConnectTimeout = 30
	}
	if service.LoadBalance.UpstreamMaxIdle == 0 {
		service.LoadBalance.UpstreamMaxIdle = 100
	}
	if service.LoadBalance.UpstreamIdleTimeout == 0 {
		service.LoadBalance.UpstreamIdleTimeout = 100
	}
	if service.LoadBalance.UpstreamHeaderTimeout == 0 {
		service.LoadBalance.UpstreamHeaderTimeout = 100
	}
	trans := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(service.LoadBalance.UpstreamConnectTimeout) * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          service.LoadBalance.UpstreamMaxIdle,
		IdleConnTimeout:       time.Duration(service.LoadBalance.UpstreamIdleTimeout) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: time.Duration(service.LoadBalance.UpstreamHeaderTimeout) * time.Second,
	}
	transItem := &TransportItem{
		Trans:       trans,
		ServiceName: service.Info.ServiceName,
	}
	t.Locker.Lock()
	defer t.Locker.Unlock()
	t.TransportSlice = append(t.TransportSlice, transItem)
	t.TransportMap[service.Info.ServiceName] = transItem
	return trans, nil
}
