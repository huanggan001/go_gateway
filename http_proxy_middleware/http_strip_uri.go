package http_proxy_middleware

import (
	"gatewat_web/dao"
	"gatewat_web/middleware"
	"gatewat_web/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

func HTTPStripUrlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		if serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL && serviceDetail.HTTPRule.NeedStripUri == 1 {
			//fmt.Println("before" + c.Request.URL.Path)
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, serviceDetail.HTTPRule.Rule, "", 1)
			//fmt.Println("before" + c.Request.URL.Path)
		}
		//http://127.0.0.1:8080/test_strip_uri/aabb
		//http://127.0.0.1:2003/aabb
		c.Next()
	}
}
