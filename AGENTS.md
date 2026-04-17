# AGENTS.md

This file provides guidance for AI agents working in this repository.

## Project Overview

GoSIP (now YSIP) is a GB28181 SIP server for video surveillance systems. It works with ZLMediaKit as the media server.
- **Go Server**: GB28181 SIP protocol with Gin HTTP API
- **Web Frontend**: Vue 3 + Vite + Ant Design Vue + Pinia


## Development
- 当用户要开发前端页面时使用 vue skill
- 当用户要开发后台时使用 golang skill

---

## Build Commands

### Go Server
```bash
# Build (Linux)
GOOS=linux go build -v -o ysip

# Build for current platform
go build -v -o ysip

# Run
go run main.go

# Format & vet (REQUIRED before committing)
go fmt ./...
go vet ./...

# Test (no test files currently exist)
go test ./...
go test -v -run "TestPattern" ./...
go test -v ./api/...

# Docker
make docker
make all
```

### Web Frontend (in web/ directory)
```bash
yarn install
yarn dev        # http://localhost:3000
yarn build
yarn lint
yarn format
```

---

## Go Swagger
```
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g main.go -o docs
```

## Code Style Guidelines

### Go
- **Formatting**: Use `gofmt` (tabs, not spaces)
- **Imports**: Group: stdlib → third-party → internal
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
- **Naming**: PascalCase for exported, camelCase for unexported. Acronyms uppercase (ID, HTTP, API)
- **Comments**: Chinese comments for exported functions
- **Error Handling**: Return early on error. Use `m.StatusXXX` constants

### API Handler Pattern
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

### Transaction Pattern
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

### Vue 3 (Composition API)
- `<script setup lang="ts">` for TypeScript
- SFC order: `<script>` → `<template>` → `<style>`
- Components: PascalCase (e.g., `UserProfile.vue`)
- Use `@` alias for `src/`: `import { useUserStore } from '@/store/user'`
- State: Pinia for global, `ref`/`reactive` for local

---

## Architecture
```
main.go → api.Init() + sipapi.Start()
              ↓                    ↓
         api/ (Gin)          sip/ (SIP Protocol)
         ├── c/ (handlers)   ├── s/ (low-level SIP)
         ├── middleware/      ├── devices.go, play.go, zlm.go
         ├── model/          └── ...
         └── service/
         db/ (GORM)          m/ (config)
         ├── gorm.go         ├── config.go, m.go
         └── tx.go           utils/
                                └── ...
```

### Frontend Structure
```
web/src/
├── components/    # 共享组件
├── views/        # 页面组件 (home, platform, devices, streams, settings)
│   └── home/     # 首页模块
│       ├── index.vue       # 首页入口
│       └── WelcomeCard.vue # 欢迎卡片组件
├── api/          # API 请求模块
├── store/        # Pinia 状态管理
├── router/       # Vue Router 配置
└── styles/       # LESS 样式 (variables.less, mixins.less)
```

### Key Directories
| Path | Description |
|------|-------------|
| `api/c/` | HTTP handlers |
| `api/middleware/` | Auth, CORS, permission |
| `api/model/` | GORM models |
| `api/service/` | Business logic |
| `sip/` | SIP protocol |
| `sip/s/` | Low-level SIP stack |
| `db/` | GORM + transactions |
| `m/` | Config, constants |
| `web/src/` | Vue 3 frontend |
| `web/src/components/` | 共享组件 |
| `web/src/views/` | 页面组件 |
| `web/src/views/home/` | 首页模块 (index.vue + WelcomeCard.vue) |

---

## API Routes

All routes use `/api/v1/` prefix:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check (no auth) |
| POST | `/api/v1/login` | Login |
| GET/POST | `/api/v1/users` | User management |
| GET/POST | `/api/v1/devices` | Device/监控管理 |
| GET/POST | `/api/v1/platform` | Platform/平台管理 (级联平台) |
| GET/POST | `/api/v1/channels` | Channel statistics only (前端无独立页面) |
| POST | `/api/v1/channels/:id/streams` | Start/stop streaming |
| GET | `/api/v1/channels/:id/records` | Recording playback |
| GET/POST | `/api/v1/roles`, `/api/v1/permissions` | RBAC |

## Frontend Pages

| Route | Component | Description |
|-------|-----------|-------------|
| `/home` | home/index.vue | 首页欢迎卡片 + WelcomeCard 组件 |
| `/platform` | platform/index.vue | 级联平台管理 |
| `/devices` | devices/index.vue | 监控设备管理 |
| `/streams` | streams/index.vue | 流管理 |
| `/settings` | settings/index.vue | 系统设置 |

**Note**: 通道(Channel)前端无独立管理页面，仅在首页统计中展示通道总数。

---

## Response Format

`m.JsonResponse(c, code, data)` produces:
```json
{"msgid": "...", "code": "0", "data": {...}}
```

Status codes:
- `m.StatusSucc` = "0" (success)
- `m.StatusAuthERR` = "1000" (auth error)
- `m.StatusDBERR` = "1001" (database error)
- `m.StatusParamsERR` = "1002" (params error)
- `m.StatusSysERR` = "1003" (system error)

---

## Database Patterns

```go
// Find with pagination and JSON filters
db.FindWithJson(db.DBClient, new(model.User), &users, filters, sort, skip, limit, true)

// Get by example
db.Get(db.DBClient, &model.User{Username: username})

// Create, Save, Delete
db.Create(db.DBClient, &user)
db.Save(db.DBClient, user)
db.Del(db.DBClient, user)

// Record not found check
if db.RecordNotFound(err) { ... }
```

---

## Notes
- **No test files exist** - test commands documented but will not produce results
- Default admin: `admin` / `admin123`
- Cron jobs every 5 min: `sipapi.CheckStreams`, `sipapi.ClearFiles`
- Config: `config.yml`
- **Logging**: Separate log files for SIP (`logs/gb28181.log`), SQL (`logs/sql.log`), and console output

---

## Logging System

### Log Files

| File | Content | Level Control |
|------|---------|---------------|
| `logs/gb28181.log` | SIP protocol interactions | Follows `logger` config |
| `logs/sql.log` | SQL queries | Enabled on `debug` or `trace` |
| Console | Application runtime logs | Follows `logger` config |

### Architecture

**SIP Logging**:
- Uses `utils.SIPLoggerHook` (function pointer injection) to avoid circular imports
- Hook is set in `main.go`: `utils.SIPLoggerHook = m.LogSIPMessage`
- `Gb28181Logger` uses `CustomFormatter` for SIP message formatting with box-drawing characters

**SQL Logging**:
- Uses custom `sqlLogWriter` (io.Writer) to intercept GORM logs
- Formats SQL logs with Location, Duration, and SQL statement
- Configured in `m/config.go`: `db.DBClient.SetLogger(log.New(GetSqlLogWriter(), "", 0))`

### Log Level Control

Configure `logger` in `config.yml`:
```yaml
logger: debug  # trace, debug, info, warn, error
```

| Level | SIP Logs | SQL Logs | App Logs |
|-------|----------|----------|----------|
| `trace` | ✅ Detailed SIP messages | ✅ All SQL queries | ✅ All |
| `debug` | ✅ SIP messages | ✅ All SQL queries | ✅ Debug+ |
| `info` | ❌ Only important info | ❌ Disabled | ✅ Info+ |
| `warn` | ❌ Only warnings | ❌ Disabled | ✅ Warn+ |
| `error` | ❌ Only errors | ❌ Disabled | ✅ Error+ |

### Key Files

| File | Purpose |
|------|---------|
| `m/logger.go` | Log initializers, formatters, level control |
| `m/config.go` | Config loading, log level setup, GORM log config |
| `utils/logger.go` | SIP log hook definition (avoids circular imports) |
| `main.go` | Hook injection (`utils.SIPLoggerHook = m.LogSIPMessage`) |

### Usage

**Log SIP Messages**:
```go
utils.LogSIPRequest(source, method, txKey, message)
utils.LogSIPResponse(source, txKey, message)
utils.LogSIPSend(msgType, destination, txKey, message)
```

**Log SQL Queries**: Automatically handled by GORM when `logger` is `debug` or `trace`.

---

## Redis Device Offline Detection

设备离线检测依赖 Redis 实现心跳保活机制。

### Redis Configuration

**config.yml:**
```yaml
redis:
  addr: localhost:6379      # Redis 地址
  password: ""              # Redis 密码（无密码留空）
  db: 0                    # Redis 数据库编号
```

**Redis Server requires:**
```
notify-keyspace-events Ex
```

### How It Works

```
设备注册/心跳(OK) → RefreshDeviceRedis() → Redis key: device:{id}, TTL=60秒
                                                  ↓
                               60秒内无新心跳 → key自动过期
                                                  ↓
                               __keyevent@0__:expired 收到过期通知
                                                  ↓
                               更新数据库 Regist=false + 发送离线通知
```

### Key Files

| File | Purpose |
|------|---------|
| `db/redis.go` | Redis 连接初始化和设备 key 操作函数 |
| `sip/handler.go` | 设备注册时写入 Redis key |
| `sip/keepalive.go` | 心跳刷新 Redis TTL + `StartDeviceOfflineWatcher` 监听过期事件 |

### Redis Key Functions

```go
// 刷新设备 Redis key（TTL 60秒）
db.RefreshDeviceRedis(deviceID)

// 删除设备 Redis key
db.DeleteDeviceRedis(deviceID)

// 检查设备是否在线
exists, _ := db.GetDeviceRedis(deviceID)
```
