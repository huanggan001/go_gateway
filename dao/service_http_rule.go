package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HttpRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"service_id" description:"服务id"`
	RuleType       int    `json:"rule_type" gorm:"rule_type" description:"匹配类型 1=域名 0=url前缀"`
	Rule           string `json:"rule" gorm:"rule" description:"type=domain表示域名(www.test.com) type=url_prefix时表示url前缀(/abc)"`
	NeedHttps      int    `json:"need_https" gorm:"need_https" description:"type=支持https 1=支持"`
	NeedWebsocket  int    `json:"need_websocket" gorm:"need_websocket" description:"启用websocket 1=启用"`
	NeedStripUri   int    `json:"need_strip_uri" gorm:"need_strip_uri" description:"启用strip_uri 1=启用"`
	UrlRewrite     string `json:"url_rewrite" gorm:"url_rewrite" description:"url重写功能，每行写一个"`
	HeaderTransfor string `json:"header_transfor" gorm:"header_transfor" description:"header转换支持增加（add）、删除(del)、修改(edit) 格式:add headname headvalue"`
}

func (rule *HttpRule) TableName() string {
	return "gateway_service_http_rule"
}

func (rule *HttpRule) Find(c *gin.Context, db *gorm.DB, search *HttpRule) (*HttpRule, error) {
	model := &HttpRule{}
	err := db.WithContext(c).Where(search).First(model).Error
	return model, err
}

func (rule *HttpRule) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.WithContext(c).Save(rule).Error; err != nil {
		return err
	}
	return nil
}

func (rule *HttpRule) ListByServiceID(c *gin.Context, db *gorm.DB, serviceID int64) ([]HttpRule, int64, error) {
	var list []HttpRule
	var count int64
	query := db.WithContext(c)
	query = query.Table(rule.TableName()).Select("*")
	query = query.Where("service_id=?", serviceID)
	err := query.Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	err = query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	return list, count, nil
}
