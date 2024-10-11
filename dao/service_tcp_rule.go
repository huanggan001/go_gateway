package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TcpRule struct {
	ID        int64 `json:"id" gorm:"primary_key"`
	ServiceID int64 `json:"service_id" gorm:"service_id" description:"服务id"`
	Port      int   `json:"port" gorm:"port" description:"端口"`
}

func (rule *TcpRule) TableName() string {
	return "gateway_service_tcp_rule"
}

func (rule *TcpRule) Find(c *gin.Context, db *gorm.DB, search *TcpRule) (*TcpRule, error) {
	model := &TcpRule{}
	err := db.WithContext(c).Where(search).First(model).Error
	return model, err
}

func (rule *TcpRule) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.WithContext(c).Save(rule).Error; err != nil {
		return err
	}
	return nil
}

func (rule *TcpRule) ListByServiceID(c *gin.Context, db *gorm.DB, serviceID int64) ([]TcpRule, int64, error) {
	var list []TcpRule
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
