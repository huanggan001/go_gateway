package dto

import (
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminSessionInfo struct {
	ID        int       `json:"id"`
	UserName  string    `json:"username"`
	LoginTime time.Time `json:"login_time"`
}

type AdminLoginInput struct {
	UserName string `json:"username" form:"username" comment:"管理员用户名" example:"admin" validate:"required"`
	PassWord string `json:"password" form:"password" comment:"管理员密码" example:"123" validate:"required"`
}

type AdminLoginOutput struct {
	Token string `json:"token" validate:""`
}

func (param *AdminLoginInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
