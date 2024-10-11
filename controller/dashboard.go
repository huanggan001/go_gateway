package controller

import (
	"gatewat_web/common/lib"
	"gatewat_web/dao"
	"gatewat_web/dto"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

type DashBoardController struct{}

func DashBoardRegister(group *gin.RouterGroup) {
	dashBoard := &DashBoardController{}
	group.GET("/panel_group_data", dashBoard.PanelGroupData)
	group.GET("/flow_stat", dashBoard.FlowStat)
	group.GET("/service_stat", dashBoard.ServiceStat)
}

// PanelGroupData godoc
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @ID /dashboard/panel_group_data
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupDataOutput} "success"
// @Router /dashboard/panel_group_data [get]
func (dashBoard *DashBoardController) PanelGroupData(c *gin.Context) {
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//从db中分页读取基本信息
	service := &dao.ServiceInfo{}
	_, ServiceTotal, err := service.PageList(c, db, &dto.ServiceListInput{PageSize: 1, PageNum: 1})
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	app := &dao.App{}
	_, AppTotal, err := app.APPList(c, db, &dto.AppListInput{PageNum: 1, PageSize: 1})
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	counter, err := public.FlowCountHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}
	output := &dto.PanelGroupDataOutput{
		ServiceNum:      ServiceTotal,
		AppNum:          AppTotal,
		TodayRequestNum: counter.TotalCount,
		CurrentQPS:      counter.QPS,
	}
	middleware.ResponseSuccess(c, output)
}

// FlowStat godoc
// @Summary 日流量统计
// @Description 日流量统计
// @Tags 首页大盘
// @ID /dashboard/flow_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /dashboard/flow_stat [get]
func (dashBoard *DashBoardController) FlowStat(c *gin.Context) {
	counter, err := public.FlowCountHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//今日流量全天小时级访问统计
	todayStat := []int64{}
	currentTime := time.Now()
	for i := 0; i <= time.Now().In(lib.TimeLocation).Hour(); i++ {
		dataTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dataTime)
		todayStat = append(todayStat, hourData)
	}

	//昨日流量全天小时级访问统计
	yesterdayStat := []int64{}
	yesterdayTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dataTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dataTime)
		yesterdayStat = append(yesterdayStat, hourData)
	}

	output := dto.StatisticsOutput{
		Today:     todayStat,
		Yesterday: yesterdayStat,
	}
	middleware.ResponseSuccess(c, output)

}

// ServiceStat godoc
// @Summary 服务类型占比
// @Description 服务类型占比
// @Tags 首页大盘
// @ID /dashboard/service_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.DashBoardServiceStatOutput} "success"
// @Router /dashboard/service_stat [get]
func (dashBoard *DashBoardController) ServiceStat(c *gin.Context) {
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	list, err := serviceInfo.GroupByLoadType(c, db)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	legend := []string{}
	for index, item := range list {
		name, ok := public.LoadTypeMap[item.LoadType]
		if !ok {
			middleware.ResponseError(c, 2003, errors.New("load_type not found"))
			return
		}
		list[index].Name = name
		legend = append(legend, name)
	}
	output := &dto.DashBoardServiceStatOutput{
		Legend: legend,
		Data:   list,
	}
	middleware.ResponseSuccess(c, output)
}
