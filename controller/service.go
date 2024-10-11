package controller

import (
	"fmt"
	"gatewat_web/common/lib"
	"gatewat_web/dao"
	"gatewat_web/dto"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type ServiceController struct{}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
	group.GET("/service_delete", service.ServiceDelete)
	group.POST("/service_add_http", service.ServiceAddHTTP)
	group.POST("/service_add_tcp", service.ServiceAddTcp)
	group.POST("/service_add_grpc", service.ServiceAddGrpc)
	group.POST("/service_update_http", service.ServiceUpdateHTTP)
	group.POST("/service_update_tcp", service.ServiceUpdateTcp)
	group.POST("/service_update_grpc", service.ServiceUpdateGrpc)
	group.GET("/service_detail", service.ServiceDetail)
	group.GET("/service_stat", service.ServiceStat)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param key_word query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_num query int true "当前页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (service *ServiceController) ServiceList(c *gin.Context) {
	params := &dto.ServiceListInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//从db中分页读取基本信息
	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.PageList(c, db, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//格式化输出信息
	outList := []dto.ServiceListItem{}
	for _, listItem := range list {
		serviceDetail, err := listItem.GetServiceDetailByServiceInfo(c, db, &listItem)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			return
		}

		//1、http后缀接入 clusterIP + clusterPort + path
		//2、http域名接入 domain
		//3、tcp、grpc接入 clusterIP + servicePort
		serviceAddr := "unknow"
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		} else if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		} else if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		} else if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		} else if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}
		ipList := serviceDetail.LoadBalance.GetIPListByModel()
		counter, err := public.FlowCountHandler.GetCounter(public.FlowServicePrefix + listItem.ServiceName)
		outItem := dto.ServiceListItem{
			ID:          listItem.ID,
			LoadType:    listItem.LoadType,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         counter.QPS,
			Qpd:         counter.TotalCount,
			TotalNode:   len(ipList),
		}
		outList = append(outList, outItem)
	}
	//格式化输出
	out := &dto.ServiceListOutput{
		Total: total,
		List:  outList,
	}
	middleware.ResponseSuccess(c, out)
}

// ServiceDetail godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/service_detail
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/service_detail [get]
func (service *ServiceController) ServiceDetail(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(c, db, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	serviceDetail, err := serviceInfo.GetServiceDetailByServiceInfo(c, db, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, serviceDetail)
}

// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务id"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_delete [get]
func (service *ServiceController) ServiceDelete(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//从db中读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(c, db, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	serviceInfo.IsDelete = 1
	if err := serviceInfo.Save(c, db); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "删除成功!")
}

// ServiceAddHTTP godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_http [post]
func (service *ServiceController) ServiceAddHTTP(c *gin.Context) {
	params := &dto.ServiceAddHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("ip列表与权重列表数量不一致"))
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	db = db.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if _, err := serviceInfo.Find(c, db, serviceInfo); err == nil {
		db.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务已存在"))
		return
	}
	httpUrl := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	if _, err := httpUrl.Find(c, db, httpUrl); err == nil {
		db.Rollback()
		middleware.ResponseError(c, 2004, errors.New("服务接入前缀或域名已存在"))
		return
	}

	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
		LoadType:    public.LoadTypeHTTP,
	}
	if err := serviceModel.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}
	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpRule.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := loadBalance.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(c, "http服务添加成功！")
}

// ServiceUpdateHTTP godoc
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @ID /service/service_update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_http [post]
func (service *ServiceController) ServiceUpdateHTTP(c *gin.Context) {
	params := &dto.ServiceUpdateHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("ip列表与权重列表数量不一致"))
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	db = db.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo, err = serviceInfo.Find(c, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}
	serviceDetail, err := serviceInfo.GetServiceDetailByServiceInfo(c, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2004, errors.New("服务不存在"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	httpRule := serviceDetail.HTTPRule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	if err := httpRule.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadBalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadBalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadBalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadBalance.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	db.Commit()
	middleware.ResponseSuccess(c, "http服务修改成功！")
}

// ServiceAddGrpc godoc
// @Summary 添加Grpc服务
// @Description 添加Grpc服务
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_grpc [post]
func (service *ServiceController) ServiceAddGrpc(c *gin.Context) {
	params := &dto.ServiceAddGrpcInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, db, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err := tcpRuleSearch.Find(c, db, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(c, db, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	db = db.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      info.ID,
		Port:           params.Port,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := grpcRule.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(c, "服务添加成功！")
}

// ServiceUpdateGrpc godoc
// @Summary 修改grpc服务
// @Description 修改grpc服务
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_grpc [post]
func (service *ServiceController) ServiceUpdateGrpc(c *gin.Context) {
	params := &dto.ServiceUpdateGrpcInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("ip列表与权重列表数量不一致"))
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	db = db.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo, err = serviceInfo.Find(c, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}
	serviceDetail, err := serviceInfo.GetServiceDetailByServiceInfo(c, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2004, errors.New("服务不存在"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	grpcRule := serviceDetail.GRPCRule
	grpcRule.HeaderTransfor = params.HeaderTransfor
	grpcRule.Port = params.Port
	if err := grpcRule.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	db.Commit()
	middleware.ResponseSuccess(c, "http服务修改成功！")
}

// ServiceAddTcp godoc
// @Summary 添加Tcp服务
// @Description 添加Tcp服务
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_tcp [post]
func (service *ServiceController) ServiceAddTcp(c *gin.Context) {
	params := &dto.ServiceAddTcpInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, db, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err := tcpRuleSearch.Find(c, db, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(c, db, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	db = db.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	tcpRule := &dao.TcpRule{
		ServiceID: info.ID,
		Port:      params.Port,
	}
	if err := tcpRule.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(c, "tcp服务添加成功！")
}

// ServiceUpdateTcp godoc
// @Summary 修改tcp服务
// @Description 修改tcp服务
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_tcp [post]
func (service *ServiceController) ServiceUpdateTcp(c *gin.Context) {
	params := &dto.ServiceUpdateTcpInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("ip列表与权重列表数量不一致"))
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	db = db.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo, err = serviceInfo.Find(c, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}
	serviceDetail, err := serviceInfo.GetServiceDetailByServiceInfo(c, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2004, errors.New("服务不存在"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	tcpRule := serviceDetail.TCPRule
	tcpRule.Port = params.Port
	if err := tcpRule.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	db.Commit()
	middleware.ResponseSuccess(c, "tcp服务修改成功！")
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/service_stat
// @Accept  json
// @Produce  json
// @Param id query string true "服务id"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/service_stat [get]
func (service *ServiceController) ServiceStat(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//从db中读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(c, db, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	counter, err := public.FlowCountHandler.GetCounter(public.FlowServicePrefix + serviceInfo.ServiceName)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	todayList := []int64{}
	currentTime := time.Now()
	for i := 0; i <= currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayList = append(todayList, hourData)
	}
	yesterdayList := []int64{}
	yesterdayTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}
	//fmt.Println(todayList)
	//fmt.Println(len(todayList))
	out := &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	}
	middleware.ResponseSuccess(c, out)
}
