# QWEN.md - GoSIP (YSIP) Project Context

## 项目概述

GoSIP（现名 YSIP）是一个基于 Go 语言的 GB28181 SIP 视频监控服务器。它与 [ZLMediaKit](https://github.com/xia-chu/ZLMediaKit) 媒体服务器配合使用，提供完整的视频监控系统解决方案。

### 核心功能
- ✅ 设备注册管理（NVR/DVR/摄像头）
- ✅ 实时预览/直播
- ✅ 远程回放/录像回放
- ✅ 录像历史文件获取
- ✅ 流管理（MySQL 存储维护，服务重启不丢失流）
- ✅ 异步通知机制
- ✅ 语音对讲
- ✅ RBAC 权限管理系统

### 技术栈

**后端：**
- **语言**: Go 1.26
- **Web 框架**: Gin (HTTP API)
- **数据库**: MySQL (GORM ORM)
- **缓存**: Redis (设备离线检测)
- **SIP 协议**: 自实现 GB28181 SIP 栈
- **API 文档**: Swagger (swaggo)
- **认证**: JWT (JSON Web Token)
- **定时任务**: robfig/cron

**前端：**
- **框架**: Vue 3 (Composition API)
- **构建工具**: Vite 5
- **UI 库**: Ant Design Vue 4
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **HTTP 客户端**: Axios
- **样式**: Less

---

## 项目结构

```
gosip/
├── main.go              # 应用入口
├── config.yml           # 配置文件
├── Makefile             # 构建脚本
├── go.mod / go.sum      # Go 依赖管理
├── api/                 # HTTP API 层
│   ├── main.go          # API 路由初始化
│   ├── c/               # HTTP Handlers (Controllers)
│   ├── middleware/      # 中间件 (Auth, CORS, Permission)
│   ├── model/           # GORM 数据模型
│   └── service/         # 业务逻辑层
├── sip/                 # SIP 协议层
│   ├── s/               # 底层 SIP 栈
│   ├── devices.go       # 设备管理
│   ├── handler.go       # SIP 请求处理
│   ├── keepalive.go     # 心跳保活
│   ├── play.go          # 播放控制
│   ├── record.go        # 录像回放
│   ├── stream.go        # 流管理
│   └── zlm.go           # ZLMediaKit 集成
├── db/                  # 数据库层
│   ├── gorm.go          # GORM 配置与工具
│   ├── redis.go         # Redis 客户端与设备离线检测
│   └── tx.go            # 事务管理
├── m/                   # 配置与常量
│   ├── config.go        # 配置加载
│   └── m.go             # 状态码与工具函数
├── utils/               # 工具函数
├── docs/                # Swagger 文档
├── demo/                # 示例配置
├── web/                 # Vue 3 前端
│   ├── src/
│   │   ├── api/         # API 请求模块
│   │   ├── components/  # 共享组件
│   │   ├── views/       # 页面组件
│   │   ├── router/      # 路由配置
│   │   ├── store/       # Pinia 状态管理
│   │   └── styles/      # Less 样式
│   └── package.json
└── logs/                # 日志目录
```

---

## 构建与运行

### 后端 (Go)

```bash
# 安装依赖
go mod download

# 运行开发模式
go run main.go

# 构建 (不自动构建)
go build -v -o dist

# 构建 (Linux)
GOOS=linux go build -v -o dist

# 格式化与检查 (提交前必须执行)
go fmt ./...
go vet ./...

# 测试 (目前无测试文件)
go test ./...

# 生成 Swagger 文档
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g main.go -o docs
```

### 前端 (Vue 3)

```bash
cd web

# 安装依赖
yarn install

# 开发服务器 (http://localhost:3000)
yarn dev

# 生产构建
yarn build

# 代码检查与格式化
yarn lint
yarn format
```

### Docker

```bash
# 构建镜像
make docker

# 完整流程 (构建 + Docker + 推送)
make all
```

---

## 配置说明

配置文件：`config.yml`

### 关键配置项

```yaml
mod: release                    # 运行模式: debug / release
database:
  dialect: mysql                # 数据库类型: mysql / postgresql / sqllite
  url: ...                      # 数据库连接字符串
redis:
  addr: localhost:6379          # Redis 地址
  password: ""                  # Redis 密码
  db: 0                         # Redis 数据库编号
udp: 0.0.0.0:55060              # SIP UDP 端口
tcp: 0.0.0.0:55060              # SIP TCP 端口
api: 0.0.0.0:8090               # HTTP API 端口
secret: z9hG4bK1233983766       # API 验证密钥
jwt:
  token_expire: 7               # JWT Token 有效期（天）
media:
  restful: http://...:18080     # ZLMediaKit RESTful API 地址
  secret: ...                   # ZLMediaKit 验证密钥
stream:
  hls: 1                        # 是否开启 HLS 转码
  rtmp: 1                       # 是否开启 RTMP 转码
gb28181:
  lid: 33010000002000000001     # 系统 ID
  region: 3301000000            # 系统域
  did: 33010000001118           # 设备 ID 前缀
  cid: 33010000001318           # 通道 ID 前缀
  pwd: 123456                   # 默认设备接入密码
```

### Redis 设备离线检测

Redis 服务端需开启键过期事件通知：
```
notify-keyspace-events Ex
```

**工作原理：**
1. 设备注册/心跳时写入 Redis key (`device:{id}`)，TTL = 60秒
2. 60秒内无新心跳则 key 自动过期
3. 程序订阅 Redis 过期事件，key 过期时自动更新设备状态为离线

---

## API 接口

所有接口使用 `/api/v1/` 前缀，响应格式：

```json
{"msgid": "...", "code": "0", "data": {...}}
```

### 状态码
| 常量 | 值 | 说明 |
|------|-----|------|
| `m.StatusSucc` | "0" | 成功 |
| `m.StatusAuthERR` | "1000" | 认证错误 |
| `m.StatusDBERR` | "1001" | 数据库错误 |
| `m.StatusParamsERR` | "1002" | 参数错误 |
| `m.StatusSysERR` | "1003" | 系统错误 |

### 主要路由

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 (无需认证) |
| POST | `/api/v1/login` | 登录 |
| GET/POST | `/api/v1/users` | 用户管理 |
| GET/POST | `/api/v1/devices` | 设备管理 |
| GET/POST | `/api/v1/platform` | 级联平台管理 |
| GET/POST | `/api/v1/channels` | 通道统计 (无独立页面) |
| POST | `/api/v1/channels/:id/streams` | 开始/停止推流 |
| GET | `/api/v1/channels/:id/records` | 录像回放 |
| GET/POST | `/api/v1/roles`, `/api/v1/permissions` | RBAC 权限管理 |

### Filters 参数使用说明

`filters` 为 JSON 编码的字符串，支持复杂查询：

```json
[
  {"field_name":"userid","opertator":"=","value":"123"},
  {"field_name":"addtime","opertator":">","value":154324556}
]
```

支持 OR 查询：
```json
[
  {"field_name":"status","opertator":"=","value":true},
  {
    "or":[
      [{"field_name":"user.add","opertator":"=","value":"234"}],
      [{"field_name":"userid","opertator":"=","value":"123"}]
    ]
  }
]
```

---

## 前端页面

| 路由 | 说明 |
|------|------|
| `/home` | 首页欢迎卡片 |
| `/platform` | 级联平台管理 |
| `/devices` | 监控设备管理 |
| `/streams` | 流管理 |
| `/settings` | 系统设置 |

**注意**: 通道 (Channel) 前端无独立管理页面，仅在首页统计中展示通道总数。

---

## 开发规范

### Go 代码风格

- **格式化**: 使用 `gofmt` (制表符，非空格)
- **导入分组**: 标准库 → 第三方 → 内部
  ```go
  import (
      "strconv"
      "github.com/gin-gonic/gin"
      "github.com/panjjo/gosip/api/model"
      "github.com/panjjo/gosip/db"
  )
  ```
- **命名**: 导出用 PascalCase，未导出用 camelCase。缩写全大写 (ID, HTTP, API)
- **注释**: 导出函数使用中文注释
- **错误处理**: 提前返回。使用 `m.StatusXXX` 常量

### API Handler 模式

```go
func HandlerName(c *gin.Context) {
    param := c.Param("id")
    if param == "" {
        m.JsonResponse(c, m.StatusParamsERR, "param cannot be empty")
        return
    }
    id, err := strconv.Atoi(param)
    if err != nil {
        m.JsonResponse(c, m.StatusParamsERR, "param format error")
        return
    }
    record := &model.Entity{}
    if err := db.DBClient.First(record, uint(id)).Error; err != nil {
        if db.RecordNotFound(err) {
            m.JsonResponse(c, m.StatusParamsERR, "record not found")
            return
        }
        m.JsonResponse(c, m.StatusDBERR, err)
        return
    }
    m.JsonResponse(c, m.StatusSucc, record)
}
```

### 事务模式

```go
tx, err := db.NewTx(db.DBClient)
if err != nil {
    m.JsonResponse(c, m.StatusDBERR, err)
    return
}
defer tx.End()
if err := db.Create(tx.DB(), &user); err != nil {
    m.JsonResponse(c, m.StatusDBERR, err)
    return
}
tx.Commit()
m.JsonResponse(c, m.StatusSucc, user)
```

### Vue 3 规范

- 使用 `<script setup lang="ts">` (Composition API)
- SFC 顺序: `<script>` → `<template>` → `<style>`
- 组件名: PascalCase (如 `UserProfile.vue`)
- 使用 `@` 别名引用 `src/`: `import { useUserStore } from '@/store/user'`
- 状态管理: Pinia (全局), `ref`/`reactive` (局部)

---

## 数据库操作

```go
// 分页查询 + JSON 过滤器
db.FindWithJson(db.DBClient, new(model.User), &users, filters, sort, skip, limit, true)

// 按示例查询
db.Get(db.DBClient, &model.User{Username: username})

// 创建、保存、删除
db.Create(db.DBClient, &user)
db.Save(db.DBClient, user)
db.Del(db.DBClient, user)

// 记录未找到检查
if db.RecordNotFound(err) { ... }
```

---

## 定时任务

每 5 分钟执行：
- `sipapi.CheckStreams` - 检查并关闭过期流
- `sipapi.ClearFiles` - 清理录制文件

---

## 默认账户

- **用户名**: `admin`
- **密码**: `admin123`

---

## 关键依赖

| 依赖 | 用途 |
|------|------|
| `github.com/gin-gonic/gin` | HTTP Web 框架 |
| `github.com/jinzhu/gorm` | ORM 数据库操作 |
| `github.com/redis/go-redis/v9` | Redis 客户端 |
| `github.com/swaggo/swag` | Swagger API 文档生成 |
| `github.com/golang-jwt/jwt/v5` | JWT 认证 |
| `github.com/spf13/viper` | 配置管理 |
| `github.com/robfig/cron` | 定时任务 |
| `github.com/sirupsen/logrus` | 日志记录 |

---

## 注意事项

- ⚠️ **无测试文件** - 测试命令已记录但不会产生结果
- ⚠️ ZLMediaKit 的 webhook 必须配置为此项目的 RESTful API 地址
- ⚠️ 首次运行时会自动创建数据库表
- ⚠️ API 文档可通过 `http://localhost:8090/swagger/index.html` 访问
- ⚠️ pprof 性能分析工具运行在 `0.0.0.0:6060`

---

## 日志规范

### 日志文件分离

系统使用分离的日志文件记录不同类型的信息：

| 日志文件 | 内容 | 级别控制 |
|---------|------|---------|
| `logs/gb28181.log` | SIP 协议交互日志 | 跟随 `logger` 配置 |
| `logs/sql.log` | SQL 查询日志 | `debug` 或 `trace` 开启 |
| 控制台 | 应用运行日志 | 跟随 `logger` 配置 |

### 日志架构

```
┌─────────────────────────────────────────────────────────┐
│  m/logger.go                                            │
│  ├── Gb28181Logger (SIP 日志记录器)                     │
│  │   └── CustomFormatter (SIP 消息框线格式化)            │
│  ├── sqlLogWriter (自定义 io.Writer)                    │
│  │   └── formatGormLog() (SQL 日志格式化)                │
│  └── SetLogLevel() (根据配置设置日志级别)                 │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ 注入
                          │
┌─────────────────────────────────────────────────────────┐
│  main.go init()                                         │
│  utils.SIPLoggerHook = m.LogSIPMessage                   │
└─────────────────────────────────────────────────────────┘
                          ↓
                          │ 调用
                          │
┌─────────────────────────────────────────────────────────┐
│  utils/logger.go                                        │
│  ├── SIPLoggerHook (钩子函数指针)                        │
│  ├── LogSIPMessage() → 调用 SIPLoggerHook               │
│  ├── LogSIPRequest()                                    │
│  ├── LogSIPResponse()                                   │
│  └── LogSIPSend()                                       │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ 使用
                          │
┌─────────────────────────────────────────────────────────┐
│  m/config.go                                            │
│  ├── LoadConfig() 读取配置                               │
│  ├── SetLogLevel() 设置日志级别                          │
│  └── GetSqlLogWriter() 配置 GORM 日志写入器               │
└─────────────────────────────────────────────────────────┘
```

### 日志级别控制

在 `config.yml` 中配置 `logger` 字段：

```yaml
logger: debug  # trace, debug, info, warn, error
```

| 级别 | SIP 日志 | SQL 日志 | 应用日志 |
|------|---------|---------|---------|
| `trace` | ✅ 详细 SIP 消息 | ✅ 所有 SQL 查询 | ✅ 全部 |
| `debug` | ✅ SIP 消息 | ✅ 所有 SQL 查询 | ✅ Debug+ |
| `info` | ❌ 仅重要信息 | ❌ 关闭 | ✅ Info+ |
| `warn` | ❌ 仅警告 | ❌ 关闭 | ✅ Warn+ |
| `error` | ❌ 仅错误 | ❌ 关闭 | ✅ Error+ |

### 日志格式示例

#### SIP 日志 (logs/gb28181.log)

使用 `CustomFormatter` 自动格式化包含 SIP 消息的日志：

```
2026-04-08 10:30:15 [DEBUG] [handler.go:57] receive request from: 192.168.1.100:5060, method: MESSAGE, txKey: abc123 message: 
┌─ SIP Message ─────────────────────────────────────────────────────────────────┐
│ MESSAGE sip:server@192.168.1.1 SIP/2.0
│ From: <sip:device@192.168.1.100>
│ ...
└───────────────────────────────────────────────────────────────────────────────┘
```

#### SQL 日志 (logs/sql.log)

使用 `sqlLogWriter` 格式化 GORM 日志：

```
┌─ SQL ─────────────────────────────────────────────────────────────────┐
│ Location: 
│   C:/Users/.../gorm.go:123
│   D:/Workbase/gosip/api/middleware/auth.go:62
│ Duration: 1.2453ms
│ SQL: SELECT * FROM `users` WHERE `users`.`deltime` IS NULL AND ((username = ?)) ORDER BY `users`.`id` ASC LIMIT 1
└─────────────────────────────────────────────────────────────────────────┘
```

### 关键文件

| 文件 | 说明 |
|------|------|
| `m/logger.go` | 日志初始化器、格式化逻辑、日志级别控制 |
| `m/config.go` | 配置加载、日志级别设置、GORM 日志配置 |
| `utils/logger.go` | SIP 日志钩子定义，避免循环导入 |
| `main.go` | 钩子注入 (`utils.SIPLoggerHook = m.LogSIPMessage`) |

---

### ZLMediaKit Webhook 配置

**重要**：ZLMediaKit 的 webhook 必须配置为此项目的 RESTful API 地址，否则部分功能（如流状态回调）无法正常使用。

在 ZLMediaKit 的配置文件 (`config.ini`) 中设置：

```ini
[hook]
enable=1
timeoutSec=10
# ZLM 回调地址，指向 GoSIP 的 API 地址
on_flow_report=http://192.168.1.192:8090/zlm/webhook/on_flow_report
on_http_access=http://192.168.1.192:8090/zlm/webhook/on_http_access
on_play=http://192.168.1.192:8090/zlm/webhook/on_play
on_publish=http://192.168.1.192:8090/zlm/webhook/on_publish
on_record_mp4=http://192.168.1.192:8090/zlm/webhook/on_record_mp4
on_rtp_server_timeout=http://192.168.1.192:8090/zlm/webhook/on_rtp_server_timeout
on_rtsp_auth=http://192.168.1.192:8090/zlm/webhook/on_rtsp_auth
on_rtsp_realm=http://192.168.1.192:8090/zlm/webhook/on_rtsp_realm
on_send_rtp_stopped=http://192.168.1.192:8090/zlm/webhook/on_send_rtp_stopped
on_server_exited=http://192.168.1.192:8090/zlm/webhook/on_server_exited
on_server_started=http://192.168.1.192:8090/zlm/webhook/on_server_started
on_shell_login=http://192.168.1.192:8090/zlm/webhook/on_shell_login
on_stream_changed=http://192.168.1.192:8090/zlm/webhook/on_stream_changed
on_stream_none_reader=http://192.168.1.192:8090/zlm/webhook/on_stream_none_reader
on_stream_not_found=http://192.168.1.192:8090/zlm/webhook/on_stream_not_found
```

**关键回调接口说明**：

| 回调事件 | 用途 |
|---------|------|
| `on_server_started` | ZLM 启动时通知，更新媒体服务器状态 |
| `on_stream_changed` | 流注册/注销通知，用于流状态管理和清理 |
| `on_stream_none_reader` | 无人观看通知，自动关闭流 |
| `on_stream_not_found` | 流不存在通知，可触发重新推流 |
| `on_publish` | 推流鉴权和配置（HLS/RTMP 转码开关）|
| `on_play` | 播放鉴权 |
| `on_record_mp4` | 录制完成通知 |

---

## AI 代理开发经验

### Skill 使用指引
- **前端开发**: 使用 `vue-best-practices` skill（Composition API + `<script setup>`）
- **后端开发**: 使用 Golang 最佳实践
- **技能发现**: 使用 `find-skills` skill 查找新能力

### 关键开发注意事项

1. **通道 (Channel) 无独立页面**
   - 前端仅在首页统计中展示通道总数
   - 通道管理通过设备详情间接处理

2. **ZLMediaKit Webhook 配置**
   - ZLM 的 webhook 必须配置为此项目的 RESTful API 地址
   - 否则部分功能（如流状态回调）无法正常使用

3. **数据库初始化**
   - 首次运行时会自动创建数据库表（GORM AutoMigrate）
   - 确保数据库已创建且连接正常

4. **流管理特性**
   - 一个通道最多一个直播申请，重复请求返回同一流
   - 每次回放请求都会产生新流，需及时关闭
   - 无人观看 5 分钟后自动关闭（可在 ZLM 配置调整）

5. **录像回放注意事项**
   - 回放前必须先调用 `/records` 获取可回放时间段
   - 传入的时间必须在回放文件时间列表内
   - 录制文件过多时最多等待 10 秒

6. **Swagger 文档更新**
   - 修改 API Handler 后需重新生成 Swagger 文档
   ```bash
   swag init -g main.go -o docs
   ```

---

## 相关资源
- **ZLMediaKit Docker**: https://hub.docker.com/repository/docker/panjjo/zlmediakit
- **Swagger 文档**: http://localhost:8090/swagger/index.html
