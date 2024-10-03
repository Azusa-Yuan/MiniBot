package emojimix

import (
	emoji_map "MiniBot/plugin/emojimix/proto"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	emoji "github.com/Andrew-M-C/go.emoji"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
)

func TestUnicode(t *testing.T) {
	var r = "👁️❗22" // 一个表情符号
	fmt.Println(len(r))
	runes := []rune(r)
	// fmt.Println(isEmoji(runes[0]))
	fmt.Println(len(runes))
	fmt.Println(runes)

	i := 0
	final := emoji.ReplaceAllEmojiFunc(r, func(emoji string) string {
		i++
		fmt.Printf("%d - %s - len %d\n", i, emoji, len(emoji))
		return ""
	})
	fmt.Println(r)
	fmt.Printf("final: <%s>", final)
}

func TestGenerateMap(t *testing.T) {
	resp, err := http.DefaultClient.Get("https://raw.githubusercontent.com/xsalazar/emoji-kitchen-backend/main/app/metadata.json")
	var body []byte
	if err != nil {
		log.Error().Err(err).Msg("")
		body, _ = os.ReadFile("./metadata.json")
	} else {
		body, _ = io.ReadAll(resp.Body)
	}

	datas := gjson.ParseBytes(body).Get("data").Map()
	outer := &emoji_map.OuterMap{OuterMap: map[string]*emoji_map.InnerMap{}}

	for _, data := range datas {
		collections := data.Get("combinations").Map()
		cur := data.Get("emoji").String()
		outer.OuterMap[cur] = &emoji_map.InnerMap{InnerMap: map[string]string{}}
		for _, endData := range collections {
			// 只取第一个
			emojiFirst := endData.Array()[0]
			leftEmoji := emojiFirst.Get("leftEmoji").String()
			rightEmoji := emojiFirst.Get("rightEmoji").String()
			if cur != leftEmoji {
				leftEmoji, rightEmoji = rightEmoji, leftEmoji
			}
			outer.OuterMap[leftEmoji].InnerMap[rightEmoji] = emojiFirst.Get("gStaticUrl").String()
		}
	}
	outFile, err := os.Create(filepath.Join(dataPath, "outer_map.bin"))
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}
	defer outFile.Close()

	data, err := proto.Marshal(outer)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}
	outFile.Write(data)

}
