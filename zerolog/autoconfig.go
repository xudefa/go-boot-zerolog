// Package zerolog 提供 Zerolog 日志适配器的自动配置。
//
// 当 zerolog.enabled=true 时自动启用，从 Environment 中读取 zerolog.level、zerolog.format、
// zerolog.time-format、zerolog.output 等配置项，
// 创建并注册 Zerolog Logger Bean 到 IoC 容器中（Bean ID: zerologLogger），实现 log.Logger 接口。
package zerolog

import (
	zerologcore "github.com/xudefa/go-boot-zerolog"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/log"
)

// init 注册 Zerolog 自动配置，由 zerolog.enabled=true 条件控制。
func init() {
	boot.RegisterAutoConfig(&ZerologAutoConfiguration{},
		condition.OnProperty("zerolog.enabled", "true"),
	)
}

// ZerologAutoConfiguration Zerolog 日志适配器的自动配置。
//
// 从 Environment 中读取 zerolog.level、zerolog.format、zerolog.output 等配置项，
// 创建 Zerolog 日志适配器并注册到 IoC 容器中，实现 log.Logger 接口。
// 启用条件：zerolog.enabled=true
type ZerologAutoConfiguration struct{}

// Configure 执行自动配置逻辑，创建 ZerologAdapter 并注册为 Bean。
func (z *ZerologAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	opts := []zerologcore.ZerologOption{
		zerologcore.WithZerologLevel(log.ToLevel(env.GetString("zerolog.level", "info"))),
		zerologcore.WithZerologFormat(env.GetString("zerolog.format", "json")),
		zerologcore.WithZerologTimeFormat(env.GetString("zerolog.time-format", "2006-01-02 15:04:05")),
	}
	if output := env.GetString("zerolog.output", ""); output != "" {
		opts = append(opts, zerologcore.WithZerologOutputPath(output))
	}

	logger := zerologcore.NewZerologAdapter(opts...)

	if err := ctx.Register("zerologLogger",
		core.Bean(logger),
		core.Singleton(),
	); err != nil {
		return err
	}

	return nil
}

// 编译时检查 ZerologAdapter 是否实现了 log.Logger 接口
var _ log.Logger = (*zerologcore.ZerologAdapter)(nil)
