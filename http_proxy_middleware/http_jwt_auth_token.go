package http_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

// jwt auth token
func HTTPJwtAuthTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//decode jwt token
		//app_id 与 app_list 取得 appInfo
		//appInfo 放到 gin.Context
		token := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
		fmt.Println("token=", token)
		appMatched := false
		if token != "" {
			claims, err := public.JwtDecode(token)
			if err != nil {
				middleware.ResponseError(c, 2002, err)
				c.Abort()
				return
			}
			appList := dao.AppManagerHandler.GetAppList()
			for _, appInfo := range appList {
				if appInfo.AppID == claims.Issuer {
					c.Set("appInfo", appInfo)
					appMatched = true
					break
				}
			}
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && !appMatched {
			middleware.ResponseError(c, 2003, errors.New("not match valid app"))
			c.Abort()
			return
		}
		c.Next()
	}
}
