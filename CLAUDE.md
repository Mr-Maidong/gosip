# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoSIP (now YSIP) is a GB28181 SIP server for video surveillance systems. It provides device management, channel management, live streaming, playback, PTZ control, and voice intercom capabilities. It works with ZLMediaKit as the media server.

## Build Commands

```bash
# Build Go server
go build -v -o srv

# Run locally
go run main.go

# Docker build
make docker

# Full release build with docker push
make all
```

## Architecture

```
main.go → api.Init() + sipapi.Start()
              ↓                    ↓
         api/ (Gin)          sip/ (SIP Protocol)
         ├── c/ (handlers)   ├── s/ (low-level SIP)
         ├── middleware/     ├── devices.go
         ├── model/          ├── play.go
         └── service/        ├── zlm.go
                               └── ...
         db/ (GORM)          m/ (config)
         ├── gorm.go         ├── config.go
         ├── model.go        └── m.go
         └── ...
```

## Key Components

- **api/** - REST API layer using Gin framework
  - `c/` - HTTP handlers/controllers
  - `middleware/` - Auth, CORS, permission middleware
  - `service/` - Business logic services
  - `model/` - Data models (User, Role, Permission, etc.)
- **sip/** - SIP protocol implementation
  - `s/` - Low-level SIP stack (parser, server, connection)
  - `devices.go` - Device registration/management
  - `play.go` - Live playback and recording
  - `zlm.go` - ZLMediaKit integration
- **db/** - Database layer (GORM)
- **m/** - Configuration loading and global state

## API Routes

The API uses `/api/v1/` prefix. Key routes:
- `GET /api/v1/health` - Health check (no auth)
- `POST /api/v1/login` - Login (no auth)
- `/api/v1/devices` - Device management
- `/api/v1/channels` - Channel management
- `/api/v1/channels/:id/streams` - Start/stop streaming
- `/api/v1/channels/:id/records` - Recording playback
- `/api/v1/users`, `/api/v1/roles`, `/api/v1/permissions` - RBAC

## Configuration

All config is in `config.yml`. Key settings:
- `udp`/`tcp` - SIP server ports (default 55060)
- `api` - REST API port (default 8090)
- `media.*` - ZLMediaKit connection settings
- `gb28181.*` - System ID, region, device/channel ID prefixes

## Default Credentials

- Admin user: `admin` / `admin123` (created on first startup)

## Cron Jobs

Two scheduled tasks run every 5 minutes:
- `sipapi.CheckStreams` - Closes inactive push streams
- `sipapi.ClearFiles` - Cleans up recording files
