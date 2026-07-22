package httpapi

import (
	"net/http"
	"strconv"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/service"
	"github.com/gin-gonic/gin"
)

// AdminHandler 处理管理 API 请求
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler 创建管理 API handler
func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// RegisterAdmin 注册管理 API 路由
func (h *AdminHandler) RegisterAdmin(router *gin.Engine) {
	admin := router.Group("/admin")
	{
		// Provider 管理
		admin.GET("/providers", h.listProviders)
		admin.GET("/providers/:name", h.getProvider)

		// 路由规则管理
		admin.GET("/routing/rules", h.listRoutingRules)
		admin.GET("/routing/rules/:id", h.getRoutingRule)
		admin.POST("/routing/rules", h.createRoutingRule)
		admin.PUT("/routing/rules/:id", h.updateRoutingRule)
		admin.DELETE("/routing/rules/:id", h.deleteRoutingRule)
	}
}

// listProviders 返回所有已注册的 provider
func (h *AdminHandler) listProviders(c *gin.Context) {
	providers := h.adminService.ListProviders()
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   providers,
	})
}

// getProvider 返回指定 provider 的信息
func (h *AdminHandler) getProvider(c *gin.Context) {
	name := c.Param("name")
	provider, err := h.adminService.GetProvider(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "not_found",
				"code":    "provider_not_found",
			},
		})
		return
	}
	c.JSON(http.StatusOK, provider)
}

// listRoutingRules 返回所有路由规则
func (h *AdminHandler) listRoutingRules(c *gin.Context) {
	rules, err := h.adminService.ListRoutingRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "internal_error",
				"code":    "database_error",
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   rules,
	})
}

// getRoutingRule 返回单条路由规则
func (h *AdminHandler) getRoutingRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "invalid rule id",
				"type":    "invalid_request_error",
				"code":    "invalid_id",
			},
		})
		return
	}

	rule, err := h.adminService.GetRoutingRule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "not_found",
				"code":    "rule_not_found",
			},
		})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// createRoutingRule 创建新的路由规则
func (h *AdminHandler) createRoutingRule(c *gin.Context) {
	var req biz.RoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	rule, err := h.adminService.CreateRoutingRule(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
				"code":    "creation_failed",
			},
		})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

// updateRoutingRule 更新路由规则
func (h *AdminHandler) updateRoutingRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "invalid rule id",
				"type":    "invalid_request_error",
				"code":    "invalid_id",
			},
		})
		return
	}

	var req biz.RoutingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
				"code":    "invalid_request",
			},
		})
		return
	}

	rule, err := h.adminService.UpdateRoutingRule(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
				"code":    "update_failed",
			},
		})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// deleteRoutingRule 删除路由规则
func (h *AdminHandler) deleteRoutingRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "invalid rule id",
				"type":    "invalid_request_error",
				"code":    "invalid_id",
			},
		})
		return
	}

	if err := h.adminService.DeleteRoutingRule(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "not_found",
				"code":    "rule_not_found",
			},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

