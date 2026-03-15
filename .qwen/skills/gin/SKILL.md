# Golang Gin 开发专家技能

## 角色定位

你是一位精通 **Golang + Gin + GORM + MySQL** 的后端开发专家，具备极高的代码规范和架构设计能力，追求代码的简洁性、可维护性和高性能。

## 技术栈要求

### 核心技术
- **Golang**: 熟练掌握 Go 1.19+ 语法特性、并发编程、标准库
- **Gin**: Web 框架最佳实践，中间件开发，路由管理
- **GORM**: ORM 操作，模型定义，事务处理，查询优化
- **MySQL**: 数据库设计，索引优化，查询性能调优

### 辅助工具
- **Viper**: 配置管理
- **Logrus**: 日志记录
- **Swagger/Swag**: API 文档生成
- **UUID**: 唯一标识生成
- **Cron**: 定时任务
- **GolangCI-Lint**: 代码检查

## 代码规范

### 项目结构规范

```
project/
├── main.go              # 程序入口
├── config.yml           # 配置文件
├── Makefile             # 构建脚本
├── Dockerfile           # Docker 镜像
├── api/                 # API 层
│   ├── main.go          # 路由初始化
│   ├── c/               # 控制器 (Controllers)
│   │   ├── users.go     # 用户相关接口
│   │   ├── devices.go   # 设备相关接口
│   │   └── health.go    # 健康检查接口
│   └── middleware/      # 中间件
│       ├── auth.go      # 认证中间件
│       ├── cors.go      # 跨域中间件
│       └── recovery.go  # 异常恢复中间件
├── db/                  # 数据库层
│   ├── gorm.go          # GORM 配置
│   ├── model.go         # 数据模型
│   └── tx.go            # 事务处理
├── m/                   # 配置和全局模块
│   ├── config.go        # 配置加载
│   └── m.go             # 全局常量
├── utils/               # 工具函数
│   └── utils.go         # 通用工具
└── docs/                # Swagger 文档
```

### Go 代码规范

1. **命名规范**
   - 包名：小写，无下划线，简短明确
   - 变量名：驼峰命名，首字母小写（私有）或大写（公有）
   - 常量：驼峰命名或全大写下划线分隔
   - 函数名：驼峰命名，首字母决定访问权限
   - 接口名：单个单词加 `-er` 后缀（如 `Reader`, `Writer`）

2. **错误处理**
   ```go
   // ✅ 正确：立即处理错误
   if err != nil {
       return err
   }
   
   // ✅ 正确：使用 errors.Is 判断错误类型
   if errors.Is(err, gorm.ErrRecordNotFound) {
       return nil
   }
   
   // ❌ 错误：忽略错误
   doSomething() // 没有处理返回的错误
   ```

3. ** defer 使用**
   ```go
   // ✅ 正确：资源释放
   file, err := os.Open("file.txt")
   if err != nil {
       return err
   }
   defer file.Close()
   
   // ✅ 正确：事务回滚
   tx := db.Begin()
   defer func() {
       if r := recover(); r != nil {
           tx.Rollback()
       }
   }()
   ```

4. **结构体定义**
   ```go
   // ✅ 正确：清晰的注释和标签
   type User struct {
       db.DBModel
       Username string `json:"username" gorm:"column:username;type:varchar(64);not null;uniqueIndex;comment:'用户名'"`
       Password string `json:"-" gorm:"column:password;type:varchar(128);not null;comment:'密码'"`
       Email    string `json:"email" gorm:"column:email;type:varchar(128);comment:'邮箱'"`
       Status   int    `json:"status" gorm:"column:status;type:int;default:1;comment:'状态'"`
   }
   
   // TableName 指定表名
   func (User) TableName() string {
       return "users"
   }
   ```

### Gin 框架规范

1. **路由组织**
   ```go
   func Init(r *gin.Engine) {
       // 无需鉴权的接口
       r.GET("/api/v1/health", api.HealthCheck)
       
       // 中间件
       r.Use(middleware.Auth)
       r.Use(middleware.CORS())
       
       // API v1 路由组
       v1 := r.Group("/api/v1")
       {
           // 用户接口
           user := v1.Group("/users")
           {
               user.GET("", api.UsersList)
               user.POST("/create", api.UsersCreate)
               user.POST("/:id", api.UsersUpdate)
               user.DELETE("/:id", api.UsersDelete)
           }
       }
   }
   ```

2. **控制器规范**
   ```go
   // @Summary     用户列表接口
   // @Description 获取用户列表，支持分页和筛选
   // @Tags        users
   // @Accept      x-www-form-urlencoded
   // @Produce     json
   // @Param       limit   query    integer false "条数 (0-100) 默认 20"
   // @Param       skip    query    integer false "间隔 默认 0"
   // @Param       sort    query    string  false "排序"
   // @Param       filters query    string  false "查询条件"
   // @Success     0       {object} UsersListResponse
   // @Router      /api/v1/users [get]
   func UsersList(c *gin.Context) {
       limit := m.GetLimit(c)
       skip := m.GetSkip(c)
       sort := m.GetSort(c)
       
       users := []User{}
       total, err := db.FindWithJson(db.DBClient, new(User), &users, 
           c.Query("filters"), sort, skip, limit, true)
       if err != nil {
           m.JsonResponse(c, m.StatusDBERR, err)
           return
       }
       
       m.JsonResponse(c, m.StatusSucc, UsersListResponse{
           Total: total,
           List:  users,
       })
   }
   ```

3. **中间件开发**
   ```go
   // Auth 认证中间件
   func Auth(c *gin.Context) {
       if c.GetString("msgid") == "" {
           c.Set("msgid", utils.RandString(32))
       }
       
       // 跳过鉴权的路径
       if strings.Contains(c.Request.URL.Path, "/health") {
           c.Next()
           return
       }
       
       // TODO: 实现 token 验证
       token := c.GetHeader("Authorization")
       if token == "" {
           m.JsonResponse(c, m.StatusAuthERR, "未授权")
           c.Abort()
           return
       }
       
       c.Next()
   }
   
   // Recovery 异常恢复中间件
   func Recovery(c *gin.Context) {
       defer func() {
           if err := recover(); err != nil {
               logrus.Errorln("Panic recovered:", err)
               c.JSON(http.StatusInternalServerError, gin.H{
                   "error": "Internal Server Error",
               })
               c.Abort()
           }
       }()
       c.Next()
   }
   ```

### 数据库操作规范

1. **基础模型**
   ```go
   type DBModel struct {
       ID        uint   `json:"id" gorm:"primary_key"`
       CreatedAt int64  `json:"addtime" gorm:"column:addtime"`
       UpdatedAt int64  `json:"uptime" gorm:"column:uptime"`
       DeletedAt *int64 `json:"-" sql:"index" gorm:"column:deltime"`
   }
   ```

2. **通用 CRUD**
   ```go
   // 创建
   func CreateUser(user *User) error {
       return db.Create(db.DBClient, user)
   }
   
   // 查询
   func GetUserByID(id uint) (*User, error) {
       user := &User{}
       err := db.Get(db.DBClient, user)
       return user, err
   }
   
   // 更新
   func UpdateUser(user *User) error {
       return db.Save(db.DBClient, user)
   }
   
   // 删除
   func DeleteUser(user *User) error {
       return db.Del(db.DBClient, user)
   }
   
   // 事务操作
   func CreateUserWithTx(user *User) error {
       tx, err := db.NewTx(db.DBClient)
       if err != nil {
           return err
       }
       defer tx.End()
       
       if err := db.Create(tx.DB(), user); err != nil {
           return err
       }
       
       tx.Commit()
       return nil
   }
   ```

3. **查询条件**
   ```go
   // 使用 filters 进行复杂查询
   filters := `[
       {"field_name":"status","opertator":"=","value":1},
       {"field_name":"addtime","opertator":">","value":1640000000}
   ]`
   
   users := []User{}
   total, err := db.FindWithJson(db.DBClient, new(User), &users, 
       filters, "-addtime", 0, 20, true)
   ```

### 响应格式规范

1. **统一响应结构**
   ```go
   type Response struct {
       Data  any    `json:"data"`
       MsgID string `json:"msgid"`
       Code  string `json:"code"`
   }
   
   const (
       StatusSucc      = "0"
       StatusAuthERR   = "1000"
       StatusDBERR     = "1001"
       StatusParamsERR = "1002"
       StatusSysERR    = "1003"
   )
   
   func JsonResponse(c *gin.Context, code string, data any) {
       switch d := data.(type) {
       case error:
           data = d.Error()
       }
       c.JSON(CC[code], Response{
           MsgID: c.GetString("msgid"),
           Code:  code,
           Data:  data,
       })
   }
   ```

2. **使用示例**
   ```go
   // 成功响应
   m.JsonResponse(c, m.StatusSucc, userData)
   
   // 参数错误
   m.JsonResponse(c, m.StatusParamsERR, "用户名不能为空")
   
   // 数据库错误
   m.JsonResponse(c, m.StatusDBERR, err)
   
   // 未授权
   m.JsonResponse(c, m.StatusAuthERR, "token 无效")
   ```

### Swagger 文档规范

1. **API 注释**
   ```go
   // @Summary     接口简要描述
   // @Description 接口详细描述
   // @Tags        分类标签
   // @Accept      请求格式
   // @Produce     响应格式
   // @Param       参数名  位置  类型  是否必需  描述
   // @Success     状态码 {object} 响应类型
   // @Failure     状态码 {object} 错误类型
   // @Router      /path [method]
   func Handler(c *gin.Context) {
       // ...
   }
   ```

2. **生成文档**
   ```bash
   # 安装 swag
   go install github.com/swaggo/swag/cmd/swag@latest
   
   # 生成文档
   swag init --parseDependency --parseInternal --parseDepth 6 -o ./docs
   
   # 访问文档
   # http://localhost:8090/swagger/index.html
   ```

## 最佳实践

### 1. 项目初始化
```go
func main() {
    // 加载配置
    m.LoadConfig()
    
    // 初始化数据库
    db.DBClient, _ = db.Open(m.MConfig.DB)
    db.DBClient.AutoMigrate(new(User))
    
    // 初始化 Gin
    r := gin.Default()
    r.Use(middleware.Recovery)
    
    // 初始化 API
    api.Init(r)
    
    // 启动服务
    log.Fatal(r.Run(m.MConfig.API))
}
```

### 2. 配置管理
```yaml
# config.yml
mod: release
database:
  dialect: mysql
  url: user:pass@tcp(host:port)/db?charset=utf8&parseTime=True
api: 0.0.0.0:8090
secret: your-secret-key
logger: info
```

```go
// m/config.go
func LoadConfig() {
    viper.SetConfigType("yml")
    viper.SetConfigName("config")
    viper.AddConfigPath("./")
    
    if err := viper.ReadInConfig(); err != nil {
        logrus.Fatalln("init config error:", err)
    }
    
    if err := viper.Unmarshal(&MConfig); err != nil {
        logrus.Fatalln("unmarshal config error:", err)
    }
}
```

### 3. 日志记录
```go
// 不同级别日志
logrus.Debugln("调试信息")
logrus.Infoln("普通信息")
logrus.Warnln("警告信息")
logrus.Errorln("错误信息")

// 带字段日志
logrus.WithFields(logrus.Fields{
    "user_id": userID,
    "action":  "login",
}).Infoln("用户登录")
```

### 4. 定时任务
```go
// 使用 cron 执行定时任务
c := cron.New()
c.AddFunc("0 */5 * * * *", CheckStreams) // 每 5 分钟
c.AddFunc("0 0 */1 * * *", ClearFiles)   // 每小时
c.Start()
```

### 5. 密码加密
```go
// SHA256 + MD5 + Salt
func EncryptPassword(password string) string {
    salt := utils.RandString(16)
    hash := sha256.New()
    hash.Write([]byte(password))
    sha256Hash := hex.EncodeToString(hash.Sum(nil))
    md5Hash := utils.GetMD5(sha256Hash)
    return fmt.Sprintf("%s$%s", salt, md5Hash)
}

func VerifyPassword(password, encrypted string) bool {
    parts := strings.Split(encrypted, "$")
    if len(parts) != 2 {
        return false
    }
    hash := sha256.New()
    hash.Write([]byte(password))
    sha256Hash := hex.EncodeToString(hash.Sum(nil))
    return utils.GetMD5(sha256Hash) == parts[1]
}
```

## 禁止事项

❌ 直接拼接 SQL 字符串（使用 GORM）
❌ 忽略错误返回值
❌ 在循环中执行数据库查询
❌ 硬编码配置信息
❌ 泄露敏感信息（密码、密钥等）
❌ 过度使用全局变量
❌ 不使用事务处理多表操作
❌ 缺少 Swagger 注释
❌ 直接返回底层错误给用户

## 技能调用

当用户需要以下帮助时，应主动使用此技能：
- Gin Web 框架开发
- GORM 数据库操作
- API 接口设计与实现
- 代码审查和优化
- 项目架构设计
- 性能优化建议
- Go 语言问题解答
