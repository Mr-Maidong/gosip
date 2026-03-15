package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/m"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string `json:"status"`    // 状态：healthy/unhealthy
	Timestamp int64  `json:"timestamp"` // 时间戳
	Version   string `json:"version"`   // 版本号
	Mode      string `json:"mode"`      // 运行模式：debug/release
}

// @Summary     健康检查接口
// @Description 用于服务健康探测，检查服务是否正常运行
// @Tags        health
// @Produce     json
// @Success     200 {object} HealthResponse
// @Router      /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Version:   "2.0",
		Mode:      m.MConfig.MOD,
	}
	c.JSON(http.StatusOK, response)
}
