package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/api/model"
	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

// TokenClaims Token 声明
type TokenClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// Auth 认证中间件
func Auth(c *gin.Context) {
	if c.GetString("msgid") == "" {
		c.Set("msgid", utils.RandString(32))
	}

	// 白名单路径，跳过认证
	println("Request path:", c.Request.URL.Path)
	if IsWhitePath(c.Request.URL.Path) {
		c.Next()
		return
	}

	// 从 Header 中获取 Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		m.JsonResponse(c, m.StatusAuthERR, "未提供认证信息")
		c.Abort()
		return
	}

	// 解析 Bearer Token
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		m.JsonResponse(c, m.StatusAuthERR, "认证格式错误")
		c.Abort()
		return
	}

	tokenString := parts[1]

	// 解析 Token
	claims, err := ParseToken(tokenString)
	if err != nil {
		logrus.Errorln("Parse token error:", err)
		m.JsonResponse(c, m.StatusAuthERR, "Token 无效或已过期")
		c.Abort()
		return
	}

	// 验证用户是否存在且启用
	user := &model.User{}
	if err := db.DBClient.First(user, claims.UserID).Error; err != nil {
		logrus.Errorln("Get user error:", err)
		m.JsonResponse(c, m.StatusAuthERR, "用户不存在")
		c.Abort()
		return
	}

	if user.Status != 1 {
		m.JsonResponse(c, m.StatusAuthERR, "用户已被禁用")
		c.Abort()
		return
	}

	// 将用户信息设置到上下文中
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)

	c.Next()
}

// GenerateToken 生成 Token
func GenerateToken(userID uint, username string) string {
	token := utils.JWTCreateToken(map[string]interface{}{
		"user_id":    userID,
		"username":   username,
		"exp":        utils.GetTime() + 86400*7, // 7 天有效期
		"issuer":     "gosip",
		"not_before": utils.GetTime(),
	})
	return token
}

// ParseToken 解析 Token
func ParseToken(tokenString string) (*TokenClaims, error) {
	claims, err := utils.JWTVerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	userID := uint(claims["user_id"].(float64))
	username := claims["username"].(string)

	return &TokenClaims{
		UserID:   userID,
		Username: username,
	}, nil
}

// Login 登录接口
// @Summary      用户登录
// @Description  用户登录获取 JWT Token，Token 有效期 7 天
// @Tags         auth
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Param        username  formData  string  true  "用户名"
// @Param        password  formData  string  true  "密码"
// @Success      0  {object}  map[string]interface{}  "登录成功，返回 Token 和用户信息（access_token, username, name, role, avatar, user_id）"
// @Failure      1000  {object}  map[string]interface{}  "参数错误"
// @Failure      1001  {object}  map[string]interface{}  "用户名或密码错误"
// @Failure      1004  {object}  map[string]interface{}  "认证错误（用户被禁用）"
// @Router       /api/v1/login [post]
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		m.JsonResponse(c, m.StatusParamsERR, "用户名和密码不能为空")
		return
	}

	// 查询用户
	user := &model.User{}
	if err := db.DBClient.Where("username = ?", username).First(user).Error; err != nil {
		logrus.Warnln("Login failed, user not found:", username)
		m.JsonResponse(c, m.StatusAuthERR, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		m.JsonResponse(c, m.StatusAuthERR, "用户已被禁用")
		return
	}

	// 验证密码
	if !utils.VerifyPassword(password, user.Password) {
		logrus.Warnln("Login failed, wrong password:", username)
		m.JsonResponse(c, m.StatusAuthERR, "用户名或密码错误")
		return
	}

	// 生成 Token
	token := GenerateToken(user.ID, user.Username)

	// 更新最后登录信息
	user.LastLoginTime = utils.GetTime()
	user.LastLoginIP = c.ClientIP()
	db.Save(db.DBClient, user)

	m.JsonResponse(c, m.StatusSucc, gin.H{
		"access_token": token,
		"username":     user.Username,
		"name":         user.Name,
		"role":         user.Role,
		"avatar":       user.Avatar,
		"user_id":      user.ID,
	})
}

// Logout 登出接口（可选，客户端删除 token 即可）
// @Summary      用户登出
// @Description  用户登出，由于使用 JWT，客户端删除 token 即可，服务端无需额外操作
// @Tags         auth
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Success      0  {string}  string  "登出成功"
// @Router       /api/v1/logout [post]
func Logout(c *gin.Context) {
	// 由于使用 JWT，服务端无需保存状态，客户端删除 token 即可
	// 如果需要实现 token 黑名单，可以在此处添加逻辑
	m.JsonResponse(c, m.StatusSucc, "登出成功")
}
