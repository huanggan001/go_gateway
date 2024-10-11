package dto

import (
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
)

type TokensInput struct {
	GrantType string `json:"grant_type" form:"grant_type" comment:"授权类型" example:"client_credentials" validate:"required"` //授权类型
	Scope     string `json:"scope" form:"scope" comment:"权限范围" example:"read_write" validate:"required"`                   //权限范围
}

func (params *TokensInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type TokensOutput struct {
	AccessToken string `json:"access_token" form:"access_token"`
	Scope       string `json:"scope" form:"scope"`
	TokenType   string `json:"token_type" form:"token_type"`
	ExpiresIn   int    `json:"expires_in" form:"expires_in"`
}
