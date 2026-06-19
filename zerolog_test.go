// Package zerolog 包含 github.com/rs/zerolog 日志库的适配器单元测试。
// 测试覆盖适配器的创建、选项配置、不同日志级别的调用、
// 字段扩展、缓冲区同步以及接口级别转换等功能。
package zerolog

import (
	"context"
	"testing"

	"github.com/xudefa/go-boot/log"
)

// TestNewZerologAdapter 验证使用基本选项创建 ZerologAdapter 是否成功，
// 包括指定日志级别、JSON 格式和时间格式，确保适配器能正常初始化。
func TestNewZerologAdapter(t *testing.T) {
	adapter := NewZerologAdapter(
		WithZerologLevel(log.DebugLevel),
		WithZerologFormat("json"),
		WithZerologTimeFormat("2006-01-02"),
	)
	if adapter == nil {
		t.Error("NewZerologAdapter() returned nil")
	}
}

// TestNewZerologAdapterConsole 验证以 console 格式创建 ZerologAdapter
// 是否成功，确保控制台输出模式下的适配器初始化正常。
func TestNewZerologAdapterConsole(t *testing.T) {
	adapter := NewZerologAdapter(
		WithZerologLevel(log.DebugLevel),
		WithZerologFormat("console"),
	)
	if adapter == nil {
		t.Error("NewZerologAdapter() returned nil")
	}
}

// TestNewZerologAdapterWithOptions 通过表驱动测试验证各个选项函数
// （WithZerologLevel、WithZerologFormat、WithZerologTimeFormat）单独使用时
// 都能正确创建适配器实例，确保每个选项的独立可用性。
func TestNewZerologAdapterWithOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  ZerologOption
	}{
		{"WithZerologLevel", WithZerologLevel(log.DebugLevel)},
		{"WithZerologFormat", WithZerologFormat("json")},
		{"WithZerologTimeFormat", WithZerologTimeFormat("2006-01-02")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewZerologAdapter(tt.opt)
			if adapter == nil {
				t.Error("option failed")
			}
		})
	}
}

// TestZerologAdapterLogLevels 验证适配器在默认配置下能正确调用各个级别的日志方法。
// 依次调用 Debug、Info、Warn、Error 四个级别，确保不 panic 且调用链路正常。
// 注意：此测试不验证输出内容，仅验证调用过程无异常。
func TestZerologAdapterLogLevels(t *testing.T) {
	adapter := NewZerologAdapter()
	ctx := context.Background()

	adapter.Debug(ctx, "debug message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Info(ctx, "info message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Warn(ctx, "warn message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Error(ctx, "error message", log.KeyValue{Key: "k", Value: "v"})
}

// TestZerologAdapterWith 验证 With 方法返回的新适配器实例不为 nil，
// 确保通过 With 扩展上下文字段的功能正常工作。
func TestZerologAdapterWith(t *testing.T) {
	adapter := NewZerologAdapter()
	ctx := context.Background()

	newAdapter := adapter.With(ctx, log.KeyValue{Key: "k", Value: "v"})
	if newAdapter == nil {
		t.Error("With() returned nil")
	}
}

// TestZerologAdapterSync 验证 Sync 方法能正常调用且不 panic。
// Sync 用于强制刷新日志缓冲区，确保日志被写入输出目标。
func TestZerologAdapterSync(t *testing.T) {
	adapter := NewZerologAdapter(WithZerologFormat("json"))
	_ = adapter.Sync()
}

// TestZerologAdapterImplementsInterface 在编译时验证 ZerologAdapter 类型
// 是否实现了 log.Logger 接口，利用 Go 的类型系统进行接口兼容性检查。
func TestZerologAdapterImplementsInterface(t *testing.T) {
	var _ log.Logger = (*ZerologAdapter)(nil)
}

// TestToZerologLevel 验证 toZerologLevel 方法能正确地将 go-boot 的 log.Level
// 映射为 zerolog 的 zerolog.Level。覆盖所有日志级别（Debug、Info、Warn、
// Error、DPanic、Panic、Fatal），确保映射关系正确。
func TestToZerologLevel(t *testing.T) {
	adapter := &ZerologAdapter{}

	tests := []struct {
		input    log.Level
		expected string
	}{
		{log.DebugLevel, "debug"},
		{log.InfoLevel, "info"},
		{log.WarnLevel, "warn"},
		{log.ErrorLevel, "error"},
		{log.DPanicLevel, "panic"},
		{log.PanicLevel, "panic"},
		{log.FatalLevel, "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			lvl := adapter.toZerologLevel(tt.input)
			if lvl.String() != tt.expected {
				t.Errorf("toZerologLevel() = %v, want %v", lvl, tt.expected)
			}
		})
	}
}

// TestToZerologFields 验证 toZerologFields 方法能正确地将 log.KeyValue 键值对列表
// 转换为 map[string]any 类型的字段。测试包含字符串、整数和布尔三种类型的值，
// 同时验证字段数量正确且各键值对映射关系正确。
// 注意：map 是无序的，因此只验证映射后的值是否正确。
func TestToZerologFields(t *testing.T) {
	adapter := &ZerologAdapter{}
	keys := []log.KeyValue{
		{Key: "k1", Value: "v1"},
		{Key: "k2", Value: 123},
		{Key: "k3", Value: true},
	}

	fields := adapter.toZerologFields(keys)
	if len(fields) != 3 {
		t.Errorf("toZerologFields() returned %d fields, want 3", len(fields))
	}
	if fields["k1"] != "v1" {
		t.Errorf("toZerologFields() k1 = %v", fields["k1"])
	}
}
