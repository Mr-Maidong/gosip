# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoSIP (YSIP) is a GB28181 SIP server for video surveillance systems, working with ZLMediaKit as the media server.
- **Go Server**: GB28181 SIP protocol with Gin HTTP API on port 8090
- **Web Frontend**: Vue 3 + Vite + Ant Design Vue + Pinia

Default admin credentials: `admin` / `admin123`

## Development Commands

### Go Server
```bash
# Build
GOOS=linux go build -v -o srv   # Linux
go build -v -o srv               # Current platform

# Run
go run main.go

# Format & vet (REQUIRED before committing)
go fmt ./...
go vet ./...

# Generate Swagger docs
swag init -g main.go -o docs

# Test (no test files currently exist)
go test ./...
```

### Web Frontend (in web/ directory)
```bash
yarn install
yarn dev        # http://localhost:3000
yarn build
yarn lint
yarn format
```

### Docker
```bash
make docker
make all
```

## Architecture

```
main.go → api.Init() + sipapi.Start()
              ↓                    ↓
         api/ (Gin HTTP)      sip/ (SIP Protocol)
         ├── c/ (handlers)   ├── s/ (low-level SIP stack)
         ├── middleware/      ├── devices.go, play.go, zlm.go
         ├── model/          └── ...
         └── service/
         db/ (GORM)          m/ (config)
         └── tx.go           utils/
```

### Frontend Structure
```
web/src/
├── views/        # Pages (home, platform, devices, streams, settings)
├── api/          # Axios request module
├── store/        # Pinia state management
├── router/       # Vue Router config
└── layouts/      # BasicLayout container
```

## Key Patterns

### Go API Handler
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
    // ...
    m.JsonResponse(c, m.StatusSucc, record)
}
```

### Go Transaction
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
- Use `@` alias for `src/`: `import { useUserStore } from '@/store/user'`

## API Routes

All routes use `/api/v1/` prefix:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check (no auth) |
| POST | `/api/v1/login` | Login |
| GET/POST | `/api/v1/users` | User management |
| GET/POST | `/api/v1/devices` | Device/监控管理 |
| POST | `/api/v1/channels/:id/streams` | Start/stop streaming |
| GET | `/api/v1/channels/:id/records` | Recording playback |
| GET/POST | `/api/v1/roles`, `/api/v1/permissions` | RBAC |

Frontend pages: `/home`, `/platform`, `/devices`, `/streams`, `/settings`

## Response Format

`m.JsonResponse(c, code, data)` produces: `{"msgid": "...", "code": "0", "data": {...}}`

Status codes: `StatusSucc`="0", `StatusAuthERR`="1000", `StatusDBERR`="1001", `StatusParamsERR`="1002", `StatusSysERR`="1003"

## Database Patterns
```go
db.FindWithJson(db.DBClient, new(model.User), &users, filters, sort, skip, limit, true)
db.Get(db.DBClient, &model.User{Username: username})
db.Create(db.DBClient, &user)
db.RecordNotFound(err) // check for not found
```

## Redis Device Offline Detection

Devices use Redis keyspace notifications for heartbeat/keepalive:
- Key: `device:{id}` with TTL 60 seconds
- Redis must have `notify-keyspace-events Ex` configured
- On key expiration → update device `Regist=false` + send offline notification

Key files: `db/redis.go`, `sip/handler.go`, `sip/keepalive.go`

## Cron Jobs (every 5 minutes)
- `sipapi.CheckStreams` - Close inactive streams
- `sipapi.ClearFiles` - Clean up recording files

## Skills

- **Frontend development**: Use vue skill
- **Backend Go development**: Use golang skill

## Config

Configuration file: `config.yml`
