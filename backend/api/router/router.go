package router

import (
	"minilend/api/handler"
	"minilend/config"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) *gin.Engine {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// API版本路由组
	apiGroup := r.Group("/api/v" + config.Config.Env.Version)
	{
		// 注册位置路由
		handler.NewPositionHandler().RegisterRoutes(apiGroup)

		// 注册价格路由
		handler.PriceHander{}.RegisterRoutes(apiGroup)

	}

	return r
}
