package controller

import (
	"encoding/json"
	"gatewat_web/common/lib"
	"gatewat_web/dao"
	"gatewat_web/dto"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginController struct{}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.GET("/login_out", adminLogin.AdminLoginOut)
}

// AdminLogin godoc
// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminLogin *AdminLoginController) AdminLogin(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//1、params.UserName 取得管理员信息 adminInfo
	//2、admininfo.salt + params.password sha256 --> saltPassword
	//3、saltPassword == admininfo.password
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(c, db, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//设置Session
	SessionInfo := &dto.AdminSessionInfo{
		ID:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	session, err := json.Marshal(SessionInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	sess := sessions.Default(c)
	sess.Set(public.AdminSessionKey, string(session))
	sess.Save()

	out := &dto.AdminLoginOutput{
		Token: params.UserName,
	}
	middleware.ResponseSuccess(c, out)
}

// AdminLoginOut godoc
// @Summary 管理员登录退出
// @Description 管理员登录退出
// @Tags 管理员接口
// @ID /admin_login/login_out
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/login_out [get]
func (adminLogin *AdminLoginController) AdminLoginOut(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Delete(public.AdminSessionKey)
	sess.Save()
	middleware.ResponseSuccess(c, "退出登陆！")
}
