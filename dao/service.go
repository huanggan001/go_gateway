package dao

import (
	"errors"
	"fmt"
	"gatewat_web/common/lib"
	"gatewat_web/dto"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

var ServiceManagerHandler *ServiceManager

func init() {
	//fmt.Println("初始化")
	ServiceManagerHandler = NewServiceManager()

}

type ServiceManager struct {
	ServiceMap   map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker       sync.RWMutex
	init         sync.Once
	err          error
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

func (s *ServiceManager) GetGrpcServiceList() []*ServiceDetail {
	list := []*ServiceDetail{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == public.LoadTypeGRPC {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) GetTcpServiceList() []*ServiceDetail {
	list := []*ServiceDetail{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == public.LoadTypeTCP {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//1、前缀匹配 /abc -->serviceSlice.rule
	//2、域名匹配 www.test.com --> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.Path
	host := c.Request.Host //www.example.com:8080
	host = host[0:strings.Index(host, ":")]
	path := c.Request.URL.Path // /abc/get
	fmt.Println(len(s.ServiceSlice))
	for _, serviceItem := range s.ServiceSlice {
		if serviceItem.Info.LoadType != public.LoadTypeHTTP {
			continue
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			if serviceItem.HTTPRule.Rule == host {
				return serviceItem, nil
			}
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL {
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("not matched service")
}

// 一次把所有服务加载到内存中
func (s *ServiceManager) LoadOnce() error {
	fmt.Println("loadonce()")
	s.init.Do(func() {
		serviceInfo := &ServiceInfo{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		db, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.ServiceListInput{PageNum: 1, PageSize: 99999}
		list, _, err := serviceInfo.PageList(c, db, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			serviceDetail, err := tmpItem.GetServiceDetailByServiceInfo(c, db, &tmpItem)
			if err != nil {
				s.err = err
				return
			}
			s.ServiceMap[listItem.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
		fmt.Println("s.ServiceSlice", len(s.ServiceSlice))
	})
	return s.err
}
