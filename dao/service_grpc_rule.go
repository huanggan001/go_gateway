package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrpcRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"service_id" description:"服务id"`
	Port           int    `json:"port" gorm:"port" description:"端口"`
	HeaderTransfor string `json:"header_transfor" gorm:"header_transfor" description:"header转换文件支持增加（add）、删除(del)、修改(edit) 格式：add headname headvalue"`
}

func (rule *GrpcRule) TableName() string {
	return "gateway_service_grpc_rule"
}

func (rule *GrpcRule) Find(c *gin.Context, db *gorm.DB, search *GrpcRule) (*GrpcRule, error) {
	model := &GrpcRule{}
	err := db.WithContext(c).Where(search).First(model).Error
	return model, err
}

func (rule *GrpcRule) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.WithContext(c).Save(rule).Error; err != nil {
		return err
	}
	return nil
}

func (rule *GrpcRule) ListByServiceID(c *gin.Context, db *gorm.DB, serviceID int64) ([]GrpcRule, int64, error) {
	var list []GrpcRule
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
