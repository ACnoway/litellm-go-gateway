package httpapi

import (
	"github.com/gin-gonic/gin"
)

// NewRouter 组装 Gin 路由与全局中间件。中间件注册顺序就是执行顺序：
// panic 恢复最外层，随后写入请求 ID，再执行日志记录和 metrics 采集，最后才执行受保护接口的鉴权。
func NewRouter(handler *Handler, adminHandler *AdminHandler, gatewayAPIKey string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), requestID(), logging(), metricsMiddleware(), authorize(gatewayAPIKey))
	handler.Register(router)
	adminHandler.RegisterAdmin(router)
	return router
}
