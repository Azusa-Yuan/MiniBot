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
	pluginName   = "poke"
	dataPath     = path.GetPluginDataPath()
	lulumuPath   = filepath.Join(dataPath, "lulumu")
	dinggongPath = filepath.Join(dataPath, "dinggong")
)

func init() {
	metaData := &zero.MetaData{
		Name: pluginName,
		Help: "",
	}
	engine := zero.NewTemplate(metaData)
	engine.On("/notice/notify/poke", zero.OnlyToMe).Handle(
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
				imgData, err := os.ReadFile(filePath)
				if err != nil {
					ctx.SendError(err)
					return
				}
				ctx.SendChain(message.ImageBytes(imgData))
				return
			} else if r > 0.3 && r < 0.6 {
				files, err := os.ReadDir(dinggongPath)
				if err != nil {
					ctx.SendError(err)
					return
				}
				index := rand.IntN(len(files))
				filePath := filepath.Join(lulumuPath, files[index].Name())
				dinggongData, err := os.ReadFile(filePath)
				if err != nil {
					ctx.SendError(err)
					return
				}
				ctx.SendChain(message.RecordBytes(dinggongData))
				return
			} else {
				ctx.SendChain(message.Poke(ctx.Event.UserID))
			}
		},
	)
}
