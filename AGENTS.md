# AGENTS.md

本文件为 AI 代理在此项目中的开发指南。

---

## 项目概述

GoSIP（现名 YSIP）是一个 GB28181 SIP 服务器，用于视频监控系统，与 ZLMediaKit 媒体服务器配合使用。
- **Go 后端**：GB28181 SIP 协议 + Gin HTTP API
- **Web 前端**：Vue 3 + Vite + Ant Design Vue + Pinia

---

## 开发规范

- 开发前端页面时使用 vue skill
- 开发后端时使用 golang skill

---

## 构建命令

### Go 后端

```bash
# 构建（Linux）
GOOS=linux go build -v -o ysip

# 当前平台构建
go build -v -o ysip

# 运行
go run main.go

# 格式化和检查（提交前必须执行）
go fmt ./...
go vet ./...

# 测试（目前无测试文件）
go test ./...
go test -v -run "TestPattern" ./...
go test -v ./api/...

# Docker
make docker
make all
```

### Web 前端（位于 web/ 目录）

```bash
yarn install
yarn dev        # http://localhost:3000
yarn build
yarn lint
yarn format
```

### Go Swagger 文档

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g main.go -o docs
```

---

## 代码风格

### Go 规范

- **格式化**：使用 `gofmt`（制表符缩进）
- **导入分组**：标准库 → 第三方 → 内部包
  ```go
  import (
      "strconv"
      "github.com/gin-gonic/gin"
      "github.com/panjjo/gosip/api/model"
      "github.com/panjjo/gosip/db"
      "github.com/panjjo/gosip/m"
      "github.com/panjjo/gosip/utils"
  )
  ```
- **命名**：导出函数/类型用 PascalCase，内部用 camelCase，缩写大写（ID、HTTP、API）
- **注释**：导出函数使用中文注释
- **错误处理**：提前返回，使用 `m.StatusXXX` 常量

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

### Vue 3（组合式 API）

- 使用 `<script setup lang="ts">`
- SFC 顺序：`<script>` → `<template>` → `<style>`
- 组件命名：PascalCase（如 `UserProfile.vue`）
- 路径别名：使用 `@` 指向 `src/`
- 状态管理：Pinia 全局状态，`ref`/`reactive` 局部状态

---

## 项目架构

```
main.go → api.Init() + sipapi.Start()
              ↓                    ↓
         api/ (Gin)          sip/ (SIP 协议)
         ├── c/ (处理器)     ├── s/ (底层 SIP 栈)
         ├── middleware/     ├── devices.go, play.go, zlm.go
         ├── model/         └── ...
         └── service/
         db/ (GORM)          m/ (配置)
         ├── gorm.go         ├── config.go, m.go
         └── tx.go           utils/
                                └── ...
```

### 前端结构

```
web/src/
├── components/    # 共享组件
├── views/         # 页面组件 (home, platform, devices, streams, settings)
│   └── home/      # 首页模块
├── api/           # API 请求
├── store/         # Pinia 状态管理
├── router/        # Vue Router 配置
└── styles/        # LESS 样式
```

---

## SIP 模块指南

### 核心数据结构

| 数据结构 | 文件 | 说明 |
|---------|------|------|
| `ActiveDevices` | `sip/devices.go` | 设备缓存，`sync.Map` 存储在线设备 |
| `StreamList` | `sip/stream.go` | 流列表，管理直播/录像/对讲流 |
| `_recordList` | `sip/record.go` | 录像列表 |

### 设备状态

- `Regist = false`：从未注册
- `Regist = true && Online = true`：在线
- `Regist = true && Online = false`：已注册但离线

### 流 ID 命名规范

- 直播：`live_{device}_{channel}`
- 录像回放：`replay_{device}_{channel}_{time}`
- 对讲：`talk_{device}_{channel}`

### 设备信令/收流地址

- `SipIP`：SIP 信令通讯 IP（优先使用）
- `StreamIP`：设备接收媒体流 IP（优先使用）

---

## API 路由

所有路由使用 `/api/v1/` 前缀：

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查（无需认证） |
| POST | `/api/v1/login` | 登录 |
| GET/POST | `/api/v1/users` | 用户管理 |
| GET/POST | `/api/v1/devices` | 设备管理 |
| GET/POST | `/api/v1/platform` | 平台管理（级联平台） |
| GET/POST | `/api/v1/channels` | 通道统计（前端无独立页面） |
| POST | `/api/v1/channels/:id/streams` | 开始/停止推流 |
| GET | `/api/v1/channels/:id/records` | 录像回放 |
| GET/POST | `/api/v1/roles`, `/api/v1/permissions` | RBAC |

### 前端页面

| 路由 | 组件 | 说明 |
|------|------|------|
| `/home` | home/index.vue | 首页欢迎卡片 |
| `/platform` | platform/index.vue | 级联平台管理 |
| `/devices` | devices/index.vue | 监控设备管理 |
| `/streams` | streams/index.vue | 流管理 |
| `/settings` | settings/index.vue | 系统设置 |

---

## 响应格式

`m.JsonResponse(c, code, data)` 输出：
```json
{"msgid": "...", "code": "0", "data": {...}}
```

状态码：
- `m.StatusSucc` = "0"（成功）
- `m.StatusAuthERR` = "1000"（认证错误）
- `m.StatusDBERR` = "1001"（数据库错误）
- `m.StatusParamsERR` = "1002"（参数错误）
- `m.StatusSysERR` = "1003"（系统错误）

---

## 数据库模式

```go
// 分页 + JSON 条件查询
db.FindWithJson(db.DBClient, new(model.User), &users, filters, sort, skip, limit, true)

// 按示例查询
db.Get(db.DBClient, &model.User{Username: username})

// 创建、保存、删除
db.Create(db.DBClient, &user)
db.Save(db.DBClient, user)
db.Del(db.DBClient, user)

// 记录不存在检查
if db.RecordNotFound(err) { ... }
```

---

## 日志系统

### 日志文件

| 文件 | 内容 | 级别控制 |
|------|------|---------|
| `logs/gb28181.log` | SIP 协议交互 | 跟随 `logger` 配置 |
| `logs/sql.log` | SQL 查询 | `debug` 或 `trace` 时启用 |
| Console | 应用运行时日志 | 跟随 `logger` 配置 |

### 架构

**SIP 日志**：
- 使用 `utils.SIPLoggerHook`（函数指针注入）避免循环导入
- 在 `main.go` 设置：`utils.SIPLoggerHook = m.LogSIPMessage`
- `Gb28181Logger` 使用 `CustomFormatter` 格式化 SIP 消息

**SQL 日志**：
- 使用自定义 `sqlLogWriter`（io.Writer）拦截 GORM 日志
- 在 `m/config.go` 配置：`db.DBClient.SetLogger(log.New(GetSqlLogWriter(), "", 0))`

### 日志级别

在 `config.yml` 中配置 `logger`：
```yaml
logger: debug  # trace, debug, info, warn, error
```

| 级别 | SIP 日志 | SQL 日志 | 应用日志 |
|------|---------|---------|---------|
| `trace` | ✅ 详细 SIP 消息 | ✅ 所有 SQL | ✅ 全部 |
| `debug` | ✅ SIP 消息 | ✅ 所有 SQL | ✅ Debug+ |
| `info` | ❌ 仅重要信息 | ❌ 禁用 | ✅ Info+ |
| `warn` | ❌ 仅警告 | ❌ 禁用 | ✅ Warn+ |
| `error` | ❌ 仅错误 | ❌ 禁用 | ✅ Error+ |

---

## Redis 设备离线检测

设备离线检测依赖 Redis 实现心跳保活机制。

### 配置

**config.yml：**
```yaml
redis:
  addr: localhost:6379
  password: ""
  db: 0
```

**Redis Server 需要：**
```
notify-keyspace-events Ex
```

### 工作流程

```
设备注册/心跳(OK) → RefreshDeviceRedis() → Redis key: device:{id}, TTL=60秒
                                                  ↓
                               60秒内无新心跳 → key自动过期
                                                  ↓
                               __keyevent@0__:expired 收到过期通知
                                                  ↓
                               更新数据库 Regist=false + 发送离线通知
```

### 核心函数

```go
db.RefreshDeviceRedis(deviceID)  // 刷新设备 Redis key（TTL 60秒）
db.DeleteDeviceRedis(deviceID)    // 删除设备 Redis key
db.GetDeviceRedis(deviceID)       // 检查设备是否在线
```

---

## 代码修改规范

### 修改原则

1. **最小化修改** - 只改需要改的部分，不要顺手改动其他代码
2. **改前确认意图** - 修改前仔细理解原有代码的意图，不能想当然
3. **改后核对 diff** - 修改后立即检查 diff，确保只有预期的变更
4. **不确定时先确认** - 如果发现原有逻辑看起来不合理，先和用户确认再修改

### 常见场景注意事项

| 场景 | 注意事项 |
|------|---------|
| 修复拼写错误 | 只改拼写单词本身，不改业务逻辑 |
| 重构变量名 | 确保只改名称，不改赋值和使用方式 |
| 优化条件判断 | 确认原逻辑后再优化，不要假设原逻辑有错 |
| 修改判断条件 | 先理解原条件的作用，确认后再改 |

---

## 待办和代码审查

详细待办事项请参考：
- `TODO.md` - 功能待办和已完成的特性
- `REVIEW.md` - 代码审查报告，包含 P0 优先修复项

### P0 优先修复项

1. **并发安全**：`_serverDevices.addr.Params` 被多 goroutine 并发修改
2. **测试覆盖率**：核心逻辑（注册、播放、心跳）缺少单元测试
3. **nonce 安全**：SIP 认证 nonce 无过期时间，存在重放攻击风险

---

## 备注

- 默认管理员账号：`admin` / `admin123`
- 定时任务（每5分钟）：`sipapi.CheckStreams`、`sipapi.ClearFiles`
- 配置文件：`config.yml`
