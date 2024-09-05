// Package main ZeroBot-Plugin main file
package main

import (
	"MiniBot/cmd"
	"MiniBot/config"
	"MiniBot/service/web"
	"MiniBot/utils"
	_ "MiniBot/utils/db"
	_ "MiniBot/utils/log"
	"fmt"

	// ---------以下插件均可通过前面加 // 注释，注释后停用并不加载插件--------- //
	_ "MiniBot/plugin/asill"
	_ "MiniBot/plugin/atri"
	_ "MiniBot/plugin/dnf"
	_ "MiniBot/plugin/emojimix"
	_ "MiniBot/plugin/event"
	_ "MiniBot/plugin/genai"
	_ "MiniBot/plugin/meme"
	_ "MiniBot/plugin/monitor"
	_ "MiniBot/plugin/moyu"
	_ "MiniBot/plugin/pcr_jjc3"
	_ "MiniBot/plugin/qqwife"
	_ "MiniBot/plugin/score"

	// -----------------------以下为内置依赖，勿动------------------------ //
	zero "ZeroBot"
	"ZeroBot/message"

	"MiniBot/plugin/manager"
	// webctrl "github.com/FloatTech/zbputils/control/web"
	// -----------------------以上为内置依赖，勿动------------------------ //
)

func init() {
	manager.Initialize()
	cmd.Execute()
	config.ConfigInit()
}

func main() {
	// 帮助
	zero.OnFullMatchGroup([]string{"help", "/help", ".help", "帮助"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("\n发送lssv ， 插件列表 或 服务列表 查看 bot 开放插件\n发送\"帮助 插件\"查看插件帮助"))
		})
	// 确保所有driver都连接上，加了个全局锁
	if web.On {
		go func() {
			r := web.GetWebEngine()
			r.Run(":8888")
		}()
	}
	zero.RunAndBlock(&config.Config.Z, utils.GlobalInitMutex.Unlock)
	fmt.Println("能运行到这？")
}
