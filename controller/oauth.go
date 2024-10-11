package controller

import (
	"encoding/base64"
	"gatewat_web/common/lib"
	"gatewat_web/dao"
	"gatewat_web/dto"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type OAuthController struct{}

func OauthRegister(group *gin.RouterGroup) {
	oauth := &OAuthController{}

	group.POST("/tokens", oauth.Tokens)

}

// Tokens godoc
// @Summary 获取Token
// @Description 获取Token
// @Tags OATUH
// @ID /oauth/tokens
// @Accept  json
// @Produce  json
// @Param body body dto.TokensInput true "body"
// @Success 200 {object} middleware.Response{data=TokensOutput} "success"
// @Router /oauth/tokens [post]
func (admin *OAuthController) Tokens(c *gin.Context) {
	params := &dto.TokensInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	splits := strings.Split(c.GetHeader("Authorization"), " ")
	if len(splits) != 2 {
		middleware.ResponseError(c, 2002, errors.New("用户名或密码格式错误"))
		return
	}

	appSecret, err := base64.StdEncoding.DecodeString(splits[1])
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	//fmt.Println(string(appSecret))

	//取出app_id_secret
	//生成 app_list
	//匹配 app_id
	//基于jwt生成token
	//生成output
	parts := strings.Split(string(appSecret), ":")
	if len(parts) != 2 {
		middleware.ResponseError(c, 2004, errors.New("用户名或密码格式错误"))
		return
	}

	appList := dao.AppManagerHandler.GetAppList()
	for _, appInfo := range appList {
		if parts[0] == appInfo.AppID && parts[1] == appInfo.Secret {
			claims := jwt.StandardClaims{
				Issuer:    appInfo.AppID,
				ExpiresAt: time.Now().Add(public.JwtExpires * time.Second).In(lib.TimeLocation).Unix(),
			}
			token, err := public.JwtEncode(claims)
			if err != nil {
				middleware.ResponseError(c, 2005, err)
				return
			}
			output := &dto.TokensOutput{
				ExpiresIn:   public.JwtExpires,
				TokenType:   "Bearer",
				AccessToken: token,
				Scope:       "write_read",
			}
			middleware.ResponseSuccess(c, output)
			return
		}
	}
	middleware.ResponseError(c, 2005, errors.New("未匹配正确app信息"))
}
