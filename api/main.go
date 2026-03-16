package api

import (
	"github.com/gin-gonic/gin"
	api "github.com/panjjo/gosip/api/c"
	"github.com/panjjo/gosip/api/middleware"
)

func Init(r *gin.Engine) {
	// 健康检查接口（无需鉴权）
	r.GET("/api/v1/health", api.HealthCheck)

	// 中间件
	r.Use(middleware.Auth)
	r.Use(middleware.CORS())

	// 设备类接口
	{
		r.GET("/api/v1/devices", api.DevicesList)
		r.POST("/api/v1/devices/create", api.DevicesCreate)
		r.POST("/api/v1/devices/:id", api.DevicesUpdate)
		r.DELETE("/api/v1/devices/:id", api.DevicesDelete)
		r.POST("/api/v1/devices/ptz", api.DevicesPTZControl)
	}
	// 通道类接口
	{
		r.POST("/api/v1/devices/:id/channels", api.ChannelCreate)
		r.POST("/api/v1/devices/:id/channels_sync", api.DeviceChannelsSync)
		r.GET("/api/v1/channels", api.ChannelsList)
		r.POST("/api/v1/channels/:id", api.ChannelsUpdate)
		r.DELETE("/api/v1/channels/:id", api.ChannelsDelete)
	}
	// 播放类接口
	{
		r.GET("/api/v1/streams", api.StreamsList)
		r.POST("/api/v1/channels/:id/streams", api.Play)
		r.POST("/api/v1/channels/:id/start_talk", api.StartTalk)
		r.DELETE("/api/v1/streams/:id", api.Stop)
	}
	// 录像类
	{
		r.GET("/api/v1/channels/:id/records", api.RecordsList)
	}
	// zlm webhook
	{
		r.POST("/zlm/webhook/:method", api.ZLMWebHook)
	}
	// ZLM 流媒体服务器状态接口
	{
		r.POST("/api/v1/stream/keepalive", api.ZLMStreamKeepalive)
		r.GET("/api/v1/media/status", api.MediaServerStatus)
	}
	// 用户类接口
	{
		r.GET("/api/v1/users", api.UsersList)
		r.POST("/api/v1/users/create", api.UsersCreate)
		r.POST("/api/v1/users/:id", api.UsersUpdate)
		r.DELETE("/api/v1/users/:id", api.UsersDelete)
		r.POST("/api/v1/users/:id/enable", api.UsersEnable)
		r.POST("/api/v1/users/:id/disable", api.UsersDisable)
		r.POST("/api/v1/users/:id/password", api.UsersChangePassword)
		// 用户角色关联
		r.GET("/api/v1/users/:id/roles", api.UsersGetRoles)
		r.POST("/api/v1/users/:id/roles", api.UsersAssignRoles)
		r.DELETE("/api/v1/users/:id/roles/:roleId", api.UsersRemoveRole)
		// 用户权限关联
		r.GET("/api/v1/users/:id/permissions", api.UsersGetPermissions)
		r.GET("/api/v1/users/:id/hasPermission", api.UsersHasPermission)
	}
	// 角色类接口
	{
		r.GET("/api/v1/roles", api.RolesList)
		r.POST("/api/v1/roles/create", api.RolesCreate)
		r.POST("/api/v1/roles/:id", api.RolesUpdate)
		r.DELETE("/api/v1/roles/:id", api.RolesDelete)
		r.POST("/api/v1/roles/:id/enable", api.RolesEnable)
		r.POST("/api/v1/roles/:id/disable", api.RolesDisable)
		r.POST("/api/v1/roles/:id/permissions", api.RolesAssignPermissions)
		r.GET("/api/v1/roles/:id/permissions", api.RolesGetPermissions)
		r.DELETE("/api/v1/roles/:id/permissions/:permissionId", api.RolesRemovePermission)
	}
	// 权限类接口
	{
		r.GET("/api/v1/permissions", api.PermissionsList)
		r.GET("/api/v1/permissions/tree", api.PermissionsTree)
		r.POST("/api/v1/permissions/create", api.PermissionsCreate)
		r.POST("/api/v1/permissions/:id", api.PermissionsUpdate)
		r.DELETE("/api/v1/permissions/:id", api.PermissionsDelete)
	}
}
