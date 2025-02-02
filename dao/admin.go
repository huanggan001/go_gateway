package dao

import (
	"fmt"
	"gatewat_web/dto"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type Admin struct {
	Id       int       `json:"id" gorm:"primary_key" description:"自增主键" `
	UserName string    `json:"user_name" gorm:"column:user_name" description:"管理员用户名" `
	Salt     string    `json:"salt" gorm:"column:salt" description:"盐" `
	Password string    `json:"password" gorm:"column:password" description:"管理员密码" `
	UpdateAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间" `
	CreateAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间" `
	IsDelete int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除" `
}

// 指定表名
func (admin *Admin) TableName() string {
	return "gateway_admin"
}

func (admin *Admin) LoginCheck(c *gin.Context, db *gorm.DB, param *dto.AdminLoginInput) (*Admin, error) {

	adminInfo, err := admin.Find(c, db, (&Admin{UserName: param.UserName, IsDelete: 0}))
	if err != nil {
		return nil, errors.New("用户信息不存在")
	}
	saltPassword := public.GenSaltPassword(adminInfo.Salt, param.PassWord)
	if adminInfo.Password != saltPassword {
		return nil, errors.New("密码错误，请重新输入")
	}
	return adminInfo, nil
}

func (admin *Admin) Find(c *gin.Context, db *gorm.DB, search *Admin) (*Admin, error) {
	out := &Admin{}
	err := db.WithContext(c).Where(search).Find(out).Error
	if err != nil {
		fmt.Println("err = ", err)
		return nil, err
	}
	return out, nil
}

func (admin *Admin) Save(c *gin.Context, db *gorm.DB) error {
	return db.WithContext(c).Save(admin).Error
}
