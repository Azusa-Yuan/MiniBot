// Package main ZeroBot-Plugin main file
package main

import (
	_ "MiniBot/utils/log"
	"MiniBot/utils/resource"
	"MiniBot/utils/schedule"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MiniBot/config"
	"MiniBot/service/web"
	_ "MiniBot/utils/db"

	// ---------以下插件均可通过前面加 // 注释，注释后停用并不加载插件--------- //
	_ "MiniBot/plugin/asill"
	_ "MiniBot/plugin/atri"

	_ "MiniBot/plugin/bilibili"
	_ "MiniBot/plugin/dnf"
	_ "MiniBot/plugin/emojimix"
	_ "MiniBot/plugin/eqa"
	_ "MiniBot/plugin/fortune"
	_ "MiniBot/plugin/genai"
	_ "MiniBot/plugin/github"
	_ "MiniBot/plugin/like"
	_ "MiniBot/plugin/meme"
	_ "MiniBot/plugin/monitor"
	_ "MiniBot/plugin/moyu"

	// _ "MiniBot/plugin/music"
	_ "MiniBot/plugin/pcr_jjc3"
	_ "MiniBot/plugin/picture_package"
	_ "MiniBot/plugin/poke"
	_ "MiniBot/plugin/qqwife"
	_ "MiniBot/plugin/score"
	_ "MiniBot/plugin/sleepmanage"

	// -----------------------以下为内置依赖，勿动------------------------ //
	zero "ZeroBot"
	"ZeroBot/message"

	"MiniBot/plugin/manager"

	"github.com/rs/zerolog/log"
	// webctrl "github.com/FloatTech/zbputils/control/web"
	// -----------------------以上为内置依赖，勿动------------------------ //
)

func init() {
	manager.Initialize()
	config.ConfigInit()
	schedule.Cron.Start()
}

func main() {
	// 帮助
	zero.OnFullMatchGroup([]string{"help", "/help", ".help", "帮助"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("发送lssv , 插件列表 或 服务列表 查看 bot 开放插件\n发送\"帮助 插件\"查看插件帮助 \n 该bot因网络问题不会每次都获取最新的qq头像，若想bot获取最新头像，可以发送 刷新头像"))
		})

	zero.Run(&config.Config.Z)

	if web.On {
		go func() {
			r := web.GetWebEngine()
			r.Run("127.0.0.1:8888")
		}()
	}

	// 优雅关闭
	// 创建一个 channel 用于监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// 等待接收到退出信号
	<-quit
	log.Info().Str("name", "main").Msg("Shutting down server...")

	errChan := make(chan error, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func() {
		errChan <- resource.ResourceManager.Cleanup(ctx)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			log.Error().Str("name", "main").Err(err).Msg("")
		} else {
			log.Info().Str("name", "main").Msg("Server exited gracefully")
		}
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			log.Error().Str("name", "main").Msg("cleanup operation timed out")
		}
	}
}
