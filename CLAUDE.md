# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

LitePan 是一个云存储聚合工具，具备媒体管理能力。后端 Go 1.26，前端 Vue 3 + TypeScript SPA。前端构建产物通过 `//go:embed` 嵌入 Go 二进制文件。

## 构建与开发命令

主要通过 Docker 构建，需支持 linux/amd64 和 linux/arm64 双架构。Dockerfile 为多阶段构建（node → go → debian runtime），Go 端设置 `CGO_ENABLED=0` 纯静态编译，天然支持交叉编译。

```bash
# Docker 多架构构建并推送（需先 docker buildx create）
docker buildx build --platform linux/amd64,linux/arm64 -t <image>:<tag> --push .

# Docker 单架构构建（本地测试）
make docker-build                                    # 默认 linux/amd64
DOCKER_PLATFORM=linux/arm64 make docker-build        # 指定 arm64

# Docker Compose
make docker-up       # docker compose up -d --build
make docker-down

# 本地构建（含 FUSE 支持）
make build

# 本地构建（不含 FUSE）
make build-nofuse

# 运行全部测试
make test

# 运行单个测试
GOWORK=off go test -race -run TestFunctionName ./internal/strm/...

# Lint（golangci-lint v2）
make lint

# 安装 linter
make lint-install
```

所有 Go 命令需加 `GOWORK=off` 前缀（上层可能存在 Go workspace）。

### 前端（web/）

```bash
cd web
npm ci
npm run dev          # 开发服务器 :5173，/api 代理到 :5211
npm run build        # 生产构建 → internal/api/web（嵌入 Go 二进制）
npm run type-check   # vue-tsc 类型检查
```

开发代理目标可通过 `LITEPAN_API_PROXY` 环境变量配置（默认 `http://127.0.0.1:5211`）。

## 架构

### 分层结构与强制导入边界

Linter（`.golangci.yml` 中的 `depguard` 规则）强制约束以下导入边界：

- `internal/domain/` — 纯领域类型与错误定义。**禁止**导入任何 `internal/*`。
- `internal/driver/` — 驱动接口定义（Meta、Lister、Downloader 等）与注册中心。
- `drivers/` — 具体云存储驱动实现。**禁止**导入 `file`、`auth`、`upload`。
- `internal/store/` — SQLite 持久层（WAL 模式，纯 Go 驱动 `modernc.org/sqlite`）。迁移文件在 `internal/store/migrations/`。
- `internal/api/` — HTTP 层（chi 路由）。**禁止**直接导入 `store`。
- `internal/file/`、`internal/playback/`、`internal/upload/` — 核心功能服务。**禁止**导入 `api`。
- `pkg/` — 纯工具包（`safego`、`secretkey`、`singleflight` 等）。**禁止**依赖业务包。
- `internal/**` — **禁止**导入具体 `drivers/`，只能依赖 `internal/driver` 接口。

### 驱动系统

`drivers/` 下每个驱动遵循统一结构：`config.go`、`driver.go`、`ops.go`、`models.go`、`auth.go`、`transport.go`、`upload.go`。驱动通过 `init()` → `driver.Register()` 自注册。`drivers/all.go` 通过空白导入聚合所有驱动。

### 应用启动

入口：`cmd/litepan/main.go`。`internal/app/app.go` 中的 `App` 结构体按依赖顺序装配所有子系统：DB（含迁移）→ Store → EventBus → Cache → DriverManager → Auth → FileService → 各功能服务 → HTTP 服务器。

### 数据库

SQLite WAL 模式，22 个迁移文件在 `internal/store/migrations/`。Store 测试使用内存 SQLite（`Options{Memory: true}`）。

### 构建标签

`fuse` 构建标签控制 FUSE 文件系统支持。参见 `internal/share/fuse/compiled_fuse.go`（`//go:build fuse`）和 `manager_stub.go`（`//go:build !fuse`）。`fusemount/` 下的平台文件使用 `//go:build linux` / `//go:build !linux`。

## 配置

环境变量及默认值：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `LITEPAN_DATA_DIR` | `./data` | 数据目录 |
| `LITEPAN_STRM_DIR` | data 目录的同级目录 | STRM 输出目录 |
| `LITEPAN_DB_PATH` | `<data_dir>/litepan.db` | SQLite 数据库路径 |
| `LITEPAN_LISTEN` | `:5211` | HTTP 监听地址 |
| `LITEPAN_LOG_LEVEL` | `info` | 日志级别 |

## 约定

- 注释和日志使用中文。
- `drivers/115/` 目录已 gitignore（私有驱动）。
- 许可证：PolyForm Noncommercial 1.0.0。
