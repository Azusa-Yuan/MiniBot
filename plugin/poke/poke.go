// 基于zhenxun_bot改造
package poke

import (
	"MiniBot/utils/path"
	zero "ZeroBot"
	"ZeroBot/message"
	"math/rand/v2"
	"os"
	"path/filepath"
)

var (
	pluginName    = "poke"
	dataPath      = path.GetPluginDataPath()
	lulumuPath    = filepath.Join(dataPath, "lulumu")
	dinggongPath  = filepath.Join(dataPath, "dinggong")
	replyMessages = []string{
		"lsp你再戳？",
		"连个可爱美少女都要戳的肥宅真恶心啊。",
		"你再戳！",
		"？再戳试试？",
		"别戳了别戳了再戳就坏了555",
		"我爪巴爪巴，球球别再戳了",
		"你戳你🐎呢？！",
		"那...那里...那里不能戳...绝对...",
		"(。´・ω・)ん?",
		"有事恁叫我，别天天一个劲戳戳戳！",
		"欸很烦欸！你戳🔨呢",
		"?",
		"再戳一下试试？",
		"???",
		"正在关闭对您的所有服务...关闭成功",
		"啊呜，太舒服刚刚竟然睡着了。什么事？",
		"正在定位您的真实地址...定位成功。轰炸机已起飞",
	}
)

func init() {
	metaData := &zero.MetaData{
		Name: pluginName,
		Help: "戳机器人，有几率发图发语音",
	}
	engine := zero.NewTemplate(metaData)
	engine.On("notice/notify/poke", zero.OnlyToMe).Handle(
		func(ctx *zero.Ctx) {
			r := rand.Float64()
			if r <= 0.3 {
				files, err := os.ReadDir(lulumuPath)
				if err != nil {
					ctx.SendError(err)
					return
				}
				index := rand.IntN(len(files))
				filePath := filepath.Join(lulumuPath, files[index].Name())
				ctx.SendChain(message.ImagePath(filePath))
				return
			} else if r <= 0.6 {
				files, err := os.ReadDir(dinggongPath)
				if err != nil {
					ctx.SendError(err)
					return
				}
				index := rand.IntN(len(files))
				filePath := filepath.Join(dinggongPath, files[index].Name())
				ctx.SendChain(message.RecordPath(filePath))
				return
			} else {
				index := rand.IntN(len(replyMessages))
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(replyMessages[index]))
			}
		},
	)
}
