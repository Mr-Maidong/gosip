package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/api/service"
	"github.com/panjjo/gosip/m"
	"github.com/sirupsen/logrus"
)

// PermissionAuth 权限认证中间件
func PermissionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户 ID（从 token 或上下文中解析）
		userID, exists := c.Get("user_id")
		if !exists {
			// 如果没有 user_id，尝试从用户名获取
			username := c.GetString("username")
			if username == "" {
				// 跳过权限检查的路径
				if IsWhitePath(c.Request.URL.Path) {
					c.Next()
					return
				}
				m.JsonResponse(c, m.StatusAuthERR, "未登录")
				c.Abort()
				return
			}

			// TODO: 从数据库查询用户 ID
			// 当前简化处理，跳过
			c.Next()
			return
		}

		// 获取当前请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 跳过权限检查的路径
		if IsWhitePath(path) {
			c.Next()
			return
		}

		// 检查权限
		permissionService := &service.PermissionService{}
		hasPermission, err := permissionService.CheckPermissionByPath(userID.(uint), path, method)

		if err != nil {
			logrus.Errorln("Check permission error:", err)
			m.JsonResponse(c, m.StatusSysERR, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			logrus.Warnln("Permission denied for user:", userID, "path:", path, "method:", method)
			m.JsonResponse(c, m.StatusAuthERR, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// IsWhitePath 判断是否是白名单路径（导出供其他包使用）
func IsWhitePath(path string) bool {
	whitePaths := []string{
		"/api/v1/health",
		"/api/v1/stream/keepalive",
		"/api/v1/login",
		"/api/v1/logout",
		"/zlm/webhook/",
		"/swagger/",
	}

	for _, whitePath := range whitePaths {
		if strings.HasPrefix(path, whitePath) {
			return true
		}
	}
	return false
}

// RequirePermission 需要特定权限的中间件
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			m.JsonResponse(c, m.StatusAuthERR, "未登录")
			c.Abort()
			return
		}

		permissionService := &service.PermissionService{}
		hasPermission, err := permissionService.CheckPermission(userID.(uint), permissionCode)

		if err != nil {
			logrus.Errorln("Check permission error:", err)
			m.JsonResponse(c, m.StatusSysERR, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			logrus.Warnln("Permission denied for user:", userID, "permission:", permissionCode)
			m.JsonResponse(c, m.StatusAuthERR, "无权限执行此操作")
			c.Abort()
			return
		}

		c.Next()
	}
}
