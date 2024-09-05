package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	router = gin.New()
	On     = false
)

func LogrusLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算请求耗时
		duration := time.Since(start)

		// 获取日志信息
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// 记录日志到 Logrus
		logrus.Infof("Request completed:  %s %s - %d %s (%v ms)", method, path, statusCode, http.StatusText(statusCode), duration.Milliseconds())
	}
}

func init() {
	router.Use(LogrusLoggerMiddleware())
}

func GetWebEngine() *gin.Engine {
	On = true
	return router
}
