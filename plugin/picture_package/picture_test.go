package picturepackage

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

func TestPage(t *testing.T) {
	maxCgPageNumber, _, _ := initPageNumber()
	for i := 1; i <= maxCgPageNumber; i++ {
		err := getPicID(i, cgType)
		if err != nil {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println(cgIDList)
	fmt.Println(len(cgIDList))
}

func TestLolicon(t *testing.T) {
	encodedTag := url.QueryEscape("刻晴")
	resp, err := http.DefaultClient.Get("https://api.lolicon.app/setu/v2?tag=" + encodedTag)
	log.Error().Err(err).Msg("")
	if err == nil && resp.StatusCode == http.StatusOK {
		respData, err := io.ReadAll(resp.Body)
		log.Error().Err(err).Msg("")
		if err == nil {
			fmt.Println(gjson.ParseBytes(respData).Get("data"))
			dataArray := gjson.ParseBytes(respData).Get("data").Array()
			if len(dataArray) != 0 {
				imgData := dataArray[0]
				url := imgData.Get("urls").Get("original").String()
				fmt.Println(url)
			}
		}
	}
}
