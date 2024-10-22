package music

import (
	"ZeroBot/message"
	"fmt"
	"net/url"
	"testing"

	"github.com/FloatTech/floatbox/web"
	"github.com/tidwall/gjson"
)

func Test163(t *testing.T) {
	requestURL := "http://music.163.com/api/search/get/web?type=1&limit=1&s=" + url.QueryEscape("心做")
	data, err := web.GetData(requestURL)
	if err != nil {
		message.Text("ERROR: ", err)
		return
	}
	fmt.Println(gjson.ParseBytes(data).Get("result.songs.0.id").Int())
}
