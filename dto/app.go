package dto

import (
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"time"
)

// 服务列表的搜索：有关键词、页数、每页条数
type AppListInput struct {
	KeyWord  string `json:"key_word" form:"key_word" comment:"关键词" example:"" validate:""`              //关键词
	PageNum  int    `json:"page_num" form:"page_num" comment:"页数" example:"1" validate:"required"`      //页数
	PageSize int    `json:"page_size" form:"page_size" comment:"每页条数" example:"20" validate:"required"` //每页条数
}

func (params *AppListInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type AppDetailInput struct {
	ID int64 `json:"id" form:"id" comment:"租户ID" validate:"required"` //关键词
}

func (params *AppDetailInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type APPListOutput struct {
	List  []APPListItemOutput `json:"list" form:"list" comment:"租户列表"`
	Total int64               `json:"total" form:"total" comment:"租户总数"`
}

type APPListItemOutput struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配		"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	RealQpd   int64     `json:"real_qpd" description:"日请求量限制"`
	RealQps   int64     `json:"real_qps" description:"每秒请求量限制"`
	UpdatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	CreatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

type APPAddInput struct {
	AppID    string `json:"app_id" form:"column:app_id" comment:"租户id" validate:"required"`
	Name     string `json:"name" form:"column:name" comment:"租户名称" validate:"required"`
	Secret   string `json:"secret" form:"column:secret" comment:"密钥" validate:""`
	WhiteIPS string `json:"white_ips" form:"column:white_ips" comment:"ip白名单，支持前缀匹配"`
	Qpd      int64  `json:"qpd" form:"column:qpd" comment:"日请求量限制" validate:""`
	Qps      int64  `json:"qps" form:"column:qps" comment:"每秒请求量限制" validate:""`
}

func (params *APPAddInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type APPUpdateInput struct {
	ID       int64  `json:"id" form:"column:id" comment:"主键ID" validate:"required"`
	AppID    string `json:"app_id" form:"column:app_id" comment:"租户id" validate:""`
	Name     string `json:"name" form:"column:name" comment:"租户名称" validate:"required"`
	Secret   string `json:"secret" form:"column:secret" comment:"密钥" validate:"required"`
	WhiteIPS string `json:"white_ips" form:"column:white_ips" comment:"ip白名单，支持前缀匹配"`
	Qpd      int64  `json:"qpd" form:"column:qpd" comment:"日请求量限制"`
	Qps      int64  `json:"qps" form:"column:qps" comment:"每秒请求量限制"`
}

func (params *APPUpdateInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type StatisticsOutput struct {
	Today     []int64 `json:"today" form:"today" comment:"今日统计" validate:"required"`
	Yesterday []int64 `json:"yesterday" form:"yesterday" comment:"昨日统计" validate:"required"`
}
