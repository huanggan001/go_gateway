package controller

import (
	"errors"
	"gatewat_web/common/lib"
	"gatewat_web/dao"
	"gatewat_web/dto"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AppController struct{}

func AppRegister(group *gin.RouterGroup) {
	admin := &AppController{}
	group.GET("/app_list", admin.APPList)
	group.GET("/app_detail", admin.APPDetail)
	group.POST("/app_add", admin.APPAdd)
	group.POST("/app_update", admin.APPUpdate)
	group.GET("/app_delete", admin.APPDelete)
	group.GET("/app_stat", admin.APPStatistics)
}

// APPList godoc
// @Summary 租户列表
// @Description 租户列表
// @Tags 租户管理
// @ID /app/app_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query string true "每页多少条"
// @Param page_num query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.APPListOutput} "success"
// @Router /app/app_list [get]
func (admin *AppController) APPList(c *gin.Context) {
	params := &dto.AppListInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	app := &dao.App{}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	list, total, err := app.APPList(c, db, params)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	outputList := []dto.APPListItemOutput{}
	for _, item := range list {
		appCounter, err := public.FlowCountHandler.GetCounter(public.FlowAppPrefix + item.AppID)
		if err != nil {
			middleware.ResponseError(c, 2004, err)
			c.Abort()
			return
		}
		outputList = append(outputList, dto.APPListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			//实时的日请求量
			RealQpd: appCounter.TotalCount,
			//实时的每秒请求量
			RealQps: appCounter.QPS,
		})
	}
	output := dto.APPListOutput{
		List:  outputList,
		Total: total,
	}
	middleware.ResponseSuccess(c, output)
}

// APPDetail godoc
// @Summary 租户详情
// @Description 租户详情
// @Tags 租户管理
// @ID /app/app_detail
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.App} "success"
// @Router /app/app_detail [get]
func (admin *AppController) APPDetail(c *gin.Context) {
	params := &dto.AppDetailInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	detail, err := search.Find(c, db, search)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, detail)
}

// APPDelete godoc
// @Summary 租户删除
// @Description 租户删除
// @Tags 租户管理
// @ID /app/app_delete
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_delete [get]
func (admin *AppController) APPDelete(c *gin.Context) {
	params := &dto.AppDetailInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	detail, err := search.Find(c, db, search)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	detail.IsDelete = 1
	if err := detail.Save(c, db); err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}
	middleware.ResponseSuccess(c, "删除成功！")
}

// APPAdd godoc
// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @ID /app/app_add
// @Accept  json
// @Produce  json
// @Param body body dto.APPAddInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_add [post]
func (admin *AppController) APPAdd(c *gin.Context) {
	params := &dto.APPAddInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//验证app_id是否被占用
	search := &dao.App{
		AppID: params.AppID,
	}
	if _, err := search.Find(c, db, search); err == nil {
		middleware.ResponseError(c, 2003, errors.New("app_id已被占用"))
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	app := &dao.App{
		AppID:    params.AppID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qpd:      params.Qpd,
		Qps:      params.Qps,
	}
	if err := app.Save(c, db); err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}
	middleware.ResponseSuccess(c, "添加成功！")
}

// APPUpdate godoc
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /app/app_update
// @Accept  json
// @Produce  json
// @Param body body dto.APPUpdateInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_update [post]
func (admin *AppController) APPUpdate(c *gin.Context) {
	params := &dto.APPUpdateInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	app := &dao.App{
		ID: params.ID,
	}
	app, err = app.Find(c, db, app)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}

	app.Qpd = params.Qpd
	app.Qps = params.Qps
	app.Secret = params.Secret
	app.WhiteIPS = params.WhiteIPS
	app.Name = params.Name

	if err := app.Save(c, db); err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}
	middleware.ResponseSuccess(c, "更新成功！")
}

// APPStatistics godoc
// @Summary 租户流量统计
// @Description 租户流量统计
// @Tags 租户管理
// @ID /app/app_stat
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dto.StatisticsOutput} "success"
// @Router /app/app_stat [get]
func (admin *AppController) APPStatistics(c *gin.Context) {
	params := &dto.AppDetailInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	app := &dao.App{
		ID: params.ID,
	}
	app, err = app.Find(c, db, app)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	//今日流量全天小时级访问统计
	todayStat := []int64{}
	counter, err := public.FlowCountHandler.GetCounter(public.FlowAppPrefix + app.AppID)
	if err != nil {
		middleware.ResponseError(c, 2004, err)
		c.Abort()
		return
	}
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
