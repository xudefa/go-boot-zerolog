package zerolog

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/xudefa/go-boot/log"
)

// ZerologOption 定义 zerolog 日志适配器的配置选项，采用函数式选项模式。
// 通过传入一系列 ZerologOption 来配置 NewZerologAdapter 的行为，
// 支持链式调用和按需定制，保持与新风格的一致性。
type ZerologOption func(*ZerologAdapter)

// WithZerologLevel 设置日志级别，控制哪些级别的日志会被输出。
// 可用的级别从低到高包括：DebugLevel、InfoLevel、WarnLevel、
// ErrorLevel、DPanicLevel、PanicLevel、FatalLevel。
// 默认级别为 InfoLevel，低于该级别的日志将被静默过滤。
func WithZerologLevel(level log.Level) ZerologOption {
	return func(a *ZerologAdapter) {
		a.level = level
	}
}

// WithZerologFormat 设置日志输出格式，支持 "json"、"console" 和 "text" 三种格式。
// - "json": JSON 格式输出（默认），适合生产环境的日志收集和分析
// - "console" 或 "text": 控制台友好格式，带彩色输出，适合开发调试
func WithZerologFormat(format string) ZerologOption {
	return func(a *ZerologAdapter) {
		a.format = format
	}
}

// WithZerologTimeFormat 设置日志时间戳的格式，使用 Go 标准时间格式布局。
// 默认格式为 "2006-01-02 15:04:05"，即年月日 时分秒。
// 可通过此选项自定义为其他格式，如 "2006-01-02" 仅显示日期部分。
func WithZerologTimeFormat(timeFormat string) ZerologOption {
	return func(a *ZerologAdapter) {
		a.timeFormat = timeFormat
	}
}

// WithZerologOutput 设置日志输出的文件句柄，可以是标准输出或已打开的文件。
// 注意 zerolog 的集成直接使用 *os.File 而非更通用的 io.Writer 接口。
// 快捷文件输出请使用 WithZerologOutputPath 按路径创建。
func WithZerologOutput(output *os.File) ZerologOption {
	return func(a *ZerologAdapter) {
		a.output = output
	}
}

// WithZerologOutputPath 设置日志文件输出路径，将日志写入指定文件。
// 如果文件不存在则创建，以追加模式写入，权限为 0644。
// 内部调用 os.OpenFile 打开文件并将结果传入 WithZerologOutput。
func WithZerologOutputPath(path string) ZerologOption {
	return func(a *ZerologAdapter) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return
		}
		a.output = f
	}
}

// ZerologAdapter 是 zerolog 日志适配器，实现 log.Logger 接口。
// 它将 github.com/rs/zerolog 高性能零分配日志库适配到 go-boot 框架的
// log.Logger 抽象接口，使得 zerolog 可以作为统一的日志实现被框架使用。
// 支持 JSON/Console 输出格式、级别控制、时间戳自动附加等功能。
type ZerologAdapter struct {
	logger     zerolog.Logger
	level      log.Level
	format     string
	timeFormat string
	output     *os.File
}

// NewZerologAdapter 创建 zerolog 日志适配器实例，通过函数式选项进行配置。
// 创建流程：
//  1. 使用默认配置初始化适配器（InfoLevel、JSON 格式、标准输出）
//  2. 依次应用传入的选项函数修改配置
//  3. 根据格式选择 ConsoleWriter（console/text）或普通 Writer（json）
//  4. 自动附加时间戳并设置日志级别
func NewZerologAdapter(opts ...ZerologOption) *ZerologAdapter {
	a := &ZerologAdapter{
		level:      log.InfoLevel,
		format:     "json",
		timeFormat: "2006-01-02 15:04:05",
		output:     os.Stdout,
	}

	for _, opt := range opts {
		opt(a)
	}

	var logger zerolog.Logger
	if a.format == "console" || a.format == "text" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: a.output}).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(a.output).With().Timestamp().Logger()
	}
	logger = logger.Level(a.toZerologLevel(a.level))
	a.logger = logger
	return a
}

// toZerologLevel 将 go-boot 框架统一的日志级别 log.Level 转换为
// zerolog 内部的 zerolog.Level 类型，实现两个日志体系的桥接。
// 注意：DPanicLevel 在 zerolog 中被映射为 PanicLevel，
// 因为 zerolog 没有 DPanic 的对应概念。当遇到未知级别时默认返回 InfoLevel。
func (a *ZerologAdapter) toZerologLevel(level log.Level) zerolog.Level {
	switch level {
	case log.DebugLevel:
		return zerolog.DebugLevel
	case log.InfoLevel:
		return zerolog.InfoLevel
	case log.WarnLevel:
		return zerolog.WarnLevel
	case log.ErrorLevel:
		return zerolog.ErrorLevel
	case log.DPanicLevel:
		return zerolog.PanicLevel
	case log.PanicLevel:
		return zerolog.PanicLevel
	case log.FatalLevel:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// toZerologFields 将 go-boot 的日志键值对列表 []log.KeyValue 转换为
// map[string]any 类型，因为 zerolog 的 Fields() 方法接受该类型。
// 如果多个键值对中出现相同的 Key，后面的值会覆盖前面的值。
func (a *ZerologAdapter) toZerologFields(keys []log.KeyValue) map[string]any {
	fields := make(map[string]any)
	for _, kv := range keys {
		fields[kv.Key] = kv.Value
	}
	return fields
}

// log 是内部统一的日志记录方法，根据传入的日志级别创建对应级别的
// zerolog 事件（通过链式调用 .Debug()/.Info()/.Warn() 等），
// 附加键值对字段后输出日志消息。ctx 参数预留用于后续支持
// 日志链路追踪上下文传递。
func (a *ZerologAdapter) log(ctx context.Context, level log.Level, msg string, keys []log.KeyValue) {
	fields := a.toZerologFields(keys)

	switch level {
	case log.DebugLevel:
		a.logger.Debug().Fields(fields).Msg(msg)
	case log.InfoLevel:
		a.logger.Info().Fields(fields).Msg(msg)
	case log.WarnLevel:
		a.logger.Warn().Fields(fields).Msg(msg)
	case log.ErrorLevel:
		a.logger.Error().Fields(fields).Msg(msg)
	case log.DPanicLevel:
		a.logger.Panic().Fields(fields).Msg(msg)
	case log.PanicLevel:
		a.logger.Panic().Fields(fields).Msg(msg)
	case log.FatalLevel:
		a.logger.Fatal().Fields(fields).Msg(msg)
	default:
		a.logger.Info().Fields(fields).Msg(msg)
	}
}

// Debug 记录调试级别的日志，用于开发和排查问题时的详细信息输出。
// 在生产环境中通常会被过滤掉以减少日志量。
func (a *ZerologAdapter) Debug(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.DebugLevel, msg, keys)
}

// Info 记录信息级别的日志，用于正常的业务运行信息记录，
// 如请求处理完成、服务启动等常规事件。
func (a *ZerologAdapter) Info(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.InfoLevel, msg, keys)
}

// Warn 记录警告级别的日志，当系统出现非关键性异常时使用。
// 表示可能存在问题但服务仍能正常运行，需要关注但不需立即处理。
func (a *ZerologAdapter) Warn(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.WarnLevel, msg, keys)
}

// Error 记录错误级别的日志，用于系统出现需要关注的异常情况。
// 表示某个操作执行失败，但不影响整个应用的继续运行。
func (a *ZerologAdapter) Error(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.ErrorLevel, msg, keys)
}

// DPanic 记录致命级别的日志并触发 panic。
// 注意：由于 zerolog 没有 DPanic（仅开发环境 panic）的对应概念，
// 此方法在 zerolog 适配器中与 Panic 行为一致，均使用 zerolog.PanicLevel。
func (a *ZerologAdapter) DPanic(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.DPanicLevel, msg, keys)
}

// Panic 记录日志后调用 panic 中止当前控制流程。
// 适用于遇到了程序无法恢复的严重错误场景。
func (a *ZerologAdapter) Panic(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.PanicLevel, msg, keys)
}

// Fatal 记录日志后调用 os.Exit(1) 立即终止程序运行。
// 用于最严重的系统级错误场景，如无法连接关键依赖服务。
func (a *ZerologAdapter) Fatal(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.FatalLevel, msg, keys)
}

// Sync 刷新日志缓冲区，将缓冲中的日志数据强制刷写到输出文件。
// 在应用优雅关闭时应调用此方法确保所有日志都已持久化，
// 避免因进程退出导致日志丢失。
// 如果输出目标为 nil 则直接返回 nil，不执行任何操作。
func (a *ZerologAdapter) Sync() error {
	if a.output != nil {
		return a.output.Sync()
	}
	return nil
}

// With 返回一个新的日志记录器，该记录器会在每条日志中自动附加指定的键值对。
// 适用于需要在上下文范围内固定携带某些字段（如请求 ID、用户 ID 等）的场景。
// 返回的新适配器共享原有适配器的其他配置（级别、格式等），
// 但拥有独立扩展的字段集，互不影响。
func (a *ZerologAdapter) With(ctx context.Context, keys ...log.KeyValue) log.Logger {
	fields := a.toZerologFields(keys)
	return &ZerologAdapter{
		logger:     a.logger.With().Fields(fields).Logger(),
		level:      a.level,
		format:     a.format,
		timeFormat: a.timeFormat,
		output:     a.output,
	}
}

// GetLogger 返回底层原始的 zerolog.Logger 实例指针。
// 用于在需要直接操作 zerolog 原生 API 的场景下提供访问能力，
// 如设置钩子（Hook）、附加自定义选项等高级用法。
func (a *ZerologAdapter) GetLogger() *zerolog.Logger {
	return &a.logger
}
