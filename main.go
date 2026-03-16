package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/api"
	"github.com/panjjo/gosip/api/middleware"
	"github.com/panjjo/gosip/api/model"
	"github.com/panjjo/gosip/api/service"
	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
	sipapi "github.com/panjjo/gosip/sip"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/panjjo/gosip/docs"

	"github.com/robfig/cron"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title          GoSIP
// @version        2.0
// @description    GB28181 SIP 服务端.
// @termsOfService https://github.com/panjjo/gosip

// @contact.name  GoSIP
// @contact.url   https://github.com/panjjo/gosip
// @contact.email panjjo@vip.qq.com

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:8090
// @BasePath /

// @securityDefinitions.basic BasicAuth

func main() {
	// 显示启动横幅
	showStartupBanner()

	// pprof
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	sipapi.Start()

	// 设置 JWT 密钥（使用配置中的 secret）
	utils.SetJWTSecret(m.MConfig.Secret)

	// 初始化权限系统（数据库表迁移和默认数据）
	db.DBClient.AutoMigrate(
		new(model.Role),
		new(model.Permission),
		new(model.UserRole),
		new(model.RolePermission),
		new(model.User),
	)
	permissionService := &service.PermissionService{}
	if err := permissionService.Init(); err != nil {
		logrus.Errorln("Init permission system error:", err)
	}

	// 创建默认管理员用户（如果不存在）
	createDefaultAdminUser()

	// 根据配置设置 Gin 运行模式
	if strings.ToUpper(m.MConfig.MOD) == "RELEASE" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()
	r.Use(middleware.Recovery)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	api.Init(r)

	logrus.Fatal(r.Run(m.MConfig.API))
}

// createDefaultAdminUser 创建默认管理员用户
func createDefaultAdminUser() {
	// 检查是否已存在管理员用户
	var count int64
	db.DBClient.Model(&model.User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	// 创建默认管理员用户
	admin := &model.User{
		Username: "admin",
		Password: utils.EncryptPassword("admin123"),
		Name:     "系统管理员",
		Email:    "admin@localhost",
		Phone:    "",
		Role:     "admin",
		Status:   1,
	}

	if err := db.Create(db.DBClient, admin); err != nil {
		logrus.Errorln("Create default admin user error:", err)
	} else {
		logrus.Info("Default admin user created: username=admin, password=admin123")
	}

	// 为管理员分配 admin 角色
	adminRole, err := service.GetRoleByCode("admin")
	if err != nil {
		logrus.Errorln("Get admin role error:", err)
		return
	}

	if adminRole != nil && adminRole.ID > 0 {
		userRole := &model.UserRole{
			UserID: admin.ID,
			RoleID: adminRole.ID,
		}
		if err := db.Create(db.DBClient, userRole); err != nil {
			logrus.Errorln("Assign admin role to default user error:", err)
		}
	}
}

func init() {
	m.LoadConfig()
	_cron()
}

func _cron() {
	c := cron.New()                                 // 新建一个定时任务对象
	c.AddFunc("0 */5 * * * *", sipapi.CheckStreams) // 定时关闭推送流
	c.AddFunc("0 */5 * * * *", sipapi.ClearFiles)   // 定时清理录制文件
	c.Start()
}

func showStartupBanner() {
	banner := `║		MySIP	GB28181 SIP Server v2.0		║`
	fmt.Println(banner)
	logrus.Infoln("GoSIP Server is starting...")
}
