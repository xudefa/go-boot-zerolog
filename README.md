# go-boot-zerolog

[![Go Version](https://img.shields.io/github/go-mod/go-version/xudefa/go-boot-zerolog)](https://go.dev/) [![License](https://img.shields.io/github/license/xudefa/go-boot-zerolog)](./LICENSE) [![Build Status](https://img.shields.io/github/actions/workflow/status/xudefa/go-boot-zerolog/test.yml?branch=master)](https://github.com/xudefa/go-boot-zerolog/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/xudefa/go-boot-zerolog.svg)](https://pkg.go.dev/github.com/xudefa/go-boot-zerolog) [![Go Report Card](https://goreportcard.com/badge/github.com/xudefa/go-boot-zerolog)](https://goreportcard.com/report/github.com/xudefa/go-boot-zerolog)

基于 [go-boot](https://github.com/xudefa/go-boot) 的 Zerolog 日志集成模块。将 github.com/rs/zerolog 无缝集成到 go-boot 的 IoC 容器和自动配置体系中，提供零分配、高性能的结构化日志记录能力。

> 设计理念：遵循 go-boot 的开发规范，通过函数式选项模式和自动配置实现零代码启动 Zerolog 日志服务。

## 整体架构

```
┌───────────────────────────────────────────────────────────────────────┐
│                    go-boot ApplicationContext                         │
│  ┌───────────┐ ┌──────────────┑ ┌───────────┐ ┌───────────┐           │
│  │ Container │ │  Environment │ │ Lifecycle │ │ EventBus  │           │
│  └───────────┘ └──────────────┘ └───────────┘ └───────────┘           │
│                       ┌─────────────────────┐                         │
│                       │ AutoConfig Registry │                         │
│                       └─────────────────────┘                         │
└───────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   go-boot-zerolog Starter     │
                    │  ┌─────────────────────────┐  │
                    │  │ ZerologAdapter Bean     │  │
                    │  │ Logger Implementation   │  │
                    │  │ Output Writer           │  │
                    │  │ Console/JSON Writer     │  │
                    │  └─────────────────────────┘  │
                    └───────────────────────────────┘
```

## 目录

- [快速开始](#快速开始)
- [功能特性](#功能特性)
- [日志记录](#日志记录)
- [高级配置](#高级配置)
- [配置选项](#配置选项)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [贡献](#贡献)
- [许可证](#许可证)

## 快速开始

### 安装

```bash
# 安装核心框架
go get github.com/xudefa/go-boot

# 安装 Zerolog 集成模块
go get github.com/xudefa/go-boot-zerolog
```

### 最小示例

```go
package main

import (
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/log"
)

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-log-app"),
        boot.WithVersion("1.0.0"),
        boot.WithProperty("zerolog.enabled", "true"),
        boot.WithProperty("zerolog.format", "console"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    // 启动应用（自动配置 Zerolog 日志）
    app.Start()

    // 获取日志实例并记录日志
    logger := app.Container().Get("zerologLogger").(log.Logger)
    
    logger.Info("Application started", 
        "version", "1.0.0",
        "mode", "production")
    logger.Warn("Deprecated API usage detected")
    logger.Error("Failed to connect to database", 
        "error", "connection refused")

    // 等待终止信号
    app.WaitForSignal()
}
```

## 功能特性

| 特性 | 说明 |
|------|------|
| Zerolog 集成 | 将 Zerolog Logger 适配到 go-boot `log.Logger` 接口 |
| 自动配置 | 通过 `zerolog.enabled=true` 自动注册日志 Bean |
| 多格式输出 | 支持 JSON 格式（生产）和 Console 格式（开发） |
| 级别控制 | 支持 Debug、Info、Warn、Error 等日志级别 |
| 零分配 | Zerolog 提供零内存分配的高性能日志记录 |
| 文件输出 | 支持日志文件输出和追加模式 |
| 自动时间戳 | 自动附加时间戳到每条日志 |

## 日志记录

### 基本日志记录

```go
logger := app.Container().Get("zerologLogger").(log.Logger)

// 不同级别的日志
logger.Debug("Debug message with details", "key", "value")
logger.Info("Information message", "user", "alice")
logger.Warn("Warning message", "retry", 3)
logger.Error("Error occurred", "error", err)
```

### 结构化日志

```go
logger.Info("User login",
    "user_id", 1001,
    "username", "alice",
    "ip", "192.168.1.100",
    "duration_ms", 150)
```

### 创建独立日志实例

```go
import "github.com/xudefa/go-boot-zerolog/zerolog"

// 创建自定义日志适配器
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologLevel(log.DebugLevel),
    zerolog.WithZerologFormat("json"),
    zerolog.WithZerologOutputPath("/var/log/app.log"),
)

logger.Info("Custom logger initialized")
```

## 高级配置

### 输出格式

```go
// JSON 格式（适合生产环境日志收集）
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologFormat("json"),
)

// Console 格式（适合本地开发调试，带彩色输出）
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologFormat("console"),
)

// Text 格式（类似 Console）
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologFormat("text"),
)
```

### 文件输出

```go
// 写入日志文件
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologOutputPath("/var/log/myapp.log"),
    zerolog.WithZerologLevel(log.InfoLevel),
)

// 或使用底层 *os.File
file, _ := os.OpenFile("/var/log/myapp.log", 
    os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologOutput(file),
)
```

### 自定义时间格式

```go
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologTimeFormat("2006-01-02"), // 仅日期
    zerolog.WithZerologTimeFormat("15:04:05"),   // 仅时间
)
```

### 标准输出

```go
// 输出到标准错误
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologOutput(os.Stderr),
)

// 输出到标准输出（默认）
logger := zerolog.NewZerologAdapter(
    zerolog.WithZerologOutput(os.Stdout),
)
```

## 配置选项

通过 `boot.WithProperty()` 或配置文件设置：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `zerolog.enabled` | `false` | 是否启用 Zerolog 日志 |
| `zerolog.level` | `info` | 日志级别（debug/info/warn/error） |
| `zerolog.format` | `json` | 输出格式（json/console/text） |
| `zerolog.time-format` | `2006-01-02 15:04:05` | 时间格式 |
| `zerolog.output` | `` | 日志文件路径（空则输出到 stdout） |

### 示例配置

```yaml
# application.yml
zerolog:
  enabled: true
  level: info
  format: json
  time-format: "2006-01-02 15:04:05"
  output: /var/log/myapp.log
```

## 项目结构

```
go-boot-zerolog/
├── zerolog.go              # Zerolog 日志适配器实现
├── autoconfig.go           # 自动配置注册
├── zerolog_test.go         # 单元测试
├── README.md
├── LICENSE
└── go.mod
```

## 开发指南

### 构建

```bash
go build ./...
```

### 测试

```bash
go test ./...
go test -cover ./...       # 带覆盖率
go test -race ./...        # 数据竞争检测
```

### 代码规范

```bash
go fmt ./...
golangci-lint run
```

## 贡献

欢迎提交 Issue 和 Pull Request！详细贡献指南请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证 — 详情请参阅 [LICENSE](./LICENSE) 文件。