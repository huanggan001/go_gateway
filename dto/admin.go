package dto

import (
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminInfoOutput struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}

type AdminChangePwd struct {
	PassWord string `json:"password" form:"password" comment:"管理员密码" example:"123" validate:"required"`
}

func (param *AdminChangePwd) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
