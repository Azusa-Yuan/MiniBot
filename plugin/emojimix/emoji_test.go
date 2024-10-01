package emojimix

import (
	emoji_map "MiniBot/plugin/emojimix/proto"
	"MiniBot/utils"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
)

func TestUnicode(t *testing.T) {
	var r rune = '🍾' // 一个表情符号
	fmt.Printf("Rune: %c, Unicode: %U  int %d\n []", r, r, r)
	fmt.Println(strconv.ParseInt("1f62e-200d-1f4a8", 16, 64))

}

// isEmoji 判断一个 rune 是否是表情符号

func TestMain(t *testing.T) {

	body, _ := os.ReadFile("./metadata.json")
	datas := gjson.ParseBytes(body).Get("data")
	os.WriteFile("./datas.json", utils.StringToBytes(datas.Raw), 0555)

}

func TestData(t *testing.T) {
	outer := &emoji_map.OuterMap{OuterMap: map[int64]*emoji_map.InnerMap{}}
	body, _ := os.ReadFile("./datas.json")
	datas := gjson.ParseBytes(body).Map()
	for key, data := range datas {
		keyInt64, err := strconv.ParseInt(key, 16, 64)
		if err != nil {
			continue
		}
		keyInt := int(keyInt64)
		fmt.Println(keyInt)
		collections := data.Get("combinations").Map()
		fmt.Println(len(collections))
		outer.OuterMap[keyInt64] = &emoji_map.InnerMap{InnerMap: map[int64]string{}}
		for otherKey, endData := range collections {
			otherKeyInt64, err := strconv.ParseInt(otherKey, 16, 64)
			if err != nil {
				continue
			}
			otherKeyInt := int(otherKeyInt64)
			fmt.Println(otherKeyInt)
			emojiFirst := endData.Array()[0]
			outer.OuterMap[keyInt64].InnerMap[otherKeyInt64] = emojiFirst.Get("gStaticUrl").String()
		}
	}
	outFile, err := os.Create("outer_map.bin")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	data, err := proto.Marshal(outer)
	if err != nil {
		fmt.Println("Error marshaling to protobuf:", err)
		return
	}
	outFile.Write(data)
	fmt.Println(len(datas))
}
