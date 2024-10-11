package controller

import (
	"encoding/json"
	"fmt"
	"gatewat_web/common/lib"
	"gatewat_web/dao"
	"gatewat_web/dto"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct{}

func AdminRegister(group *gin.RouterGroup) {
	adminLogin := &AdminController{}
	group.GET("/admin_info", adminLogin.AdminInfo)
	group.POST("/changePwd", adminLogin.AdminChangePwd)
}

// AdminInfo godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (admin *AdminController) AdminInfo(c *gin.Context) {
	session := sessions.Default(c)
	sessionInfo := session.Get(public.AdminSessionKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessionInfo)), adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//1、读取sessionKey 对应json 转换为结构体
	//2、取出数据然后封装输出结构体

	//Avatar string `json:"avatat"`
	//Introduction string `json:"introduction"`
	//Roles []string `json:"roles"`
	out := &dto.AdminInfoOutput{
		ID:           adminSessionInfo.ID,
		Name:         adminSessionInfo.UserName,
		LoginTime:    adminSessionInfo.LoginTime,
		Avatar:       "",
		Introduction: "super adminstrator",
		Roles:        []string{"admin"},
	}
	middleware.ResponseSuccess(c, out)
}

// AdminChangePwd godoc
// @Summary 管理员密码修改
// @Description 管理员密码修改
// @Tags 管理员接口
// @ID /admin/changePwd
// @Accept  json
// @Produce  json
// @Param body body dto.AdminChangePwd true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/changePwd [post]
func (admin *AdminController) AdminChangePwd(c *gin.Context) {
	params := &dto.AdminChangePwd{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//1、从session中获取用户信息到结构体 sessInfo
	//2、sessInfo.ID 读取数据库信息 adminInfo
	//3、params.salt + newPassWord sha256 ---> saltPassWord
	session := sessions.Default(c)
	sessInfo := session.Get(public.AdminSessionKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//从数据库中读取 adminInfo
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	adminInfo := &dao.Admin{}
	adminInfo, err = adminInfo.Find(c, db, (&dao.Admin{UserName: adminSessionInfo.UserName}))
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//修改密码
	newPassword := public.GenSaltPassword(adminInfo.Salt, params.PassWord)
	adminInfo.Password = newPassword
	if err := adminInfo.Save(c, db); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "修改密码成功！")
}
