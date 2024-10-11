package http_proxy_middleware

import (
	"fmt"
	"gatewat_web/dao"
	"gatewat_web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	regexp2 "regexp"
	"strings"
)

func HTTPUrlRewriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		for _, item := range strings.Split(serviceDetail.HTTPRule.UrlRewrite, ",") {
			items := strings.Split(item, " ")
			if len(items) != 2 {
				continue
			}
			regexp, err := regexp2.Compile(items[0])
			if err != nil {
				continue
			}
			fmt.Println("before rewrite", c.Request.URL.Path)
			replacePath := regexp.ReplaceAll([]byte(c.Request.URL.Path), []byte(items[1]))
			c.Request.URL.Path = string(replacePath)
			fmt.Println("after rewrite", c.Request.URL.Path)
		}
		c.Next()
	}
}
