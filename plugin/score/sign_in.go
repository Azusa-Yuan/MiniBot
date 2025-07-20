// Package score 签到
package score

import (
	"MiniBot/utils/cache"
	"MiniBot/utils/net_tools"
	"MiniBot/utils/path"
	"MiniBot/utils/transform"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	zero "ZeroBot"

	"ZeroBot/message"

	"MiniBot/service/wallet"
	"MiniBot/utils/file"
	"MiniBot/utils/text"

	"github.com/FloatTech/imgfactory"
	"golang.org/x/image/webp"

	"github.com/golang/freetype"
	"github.com/wcharczuk/go-chart/v2"
)

const (
	backgroundURL = "https://pic.re/image"
	referer       = "https://weibo.com/"
	signinMax     = 1
	// SCOREMAX 分数上限定为1200
	SCOREMAX = 1200
)

var (
	cachePath string
	rankArray = [...]int{0, 10, 20, 50, 100, 200, 350, 550, 750, 1000, 1200}

	metaData = &zero.Metadata{
		Name: "签到",
		Help: "- 签到\n- 获得签到背景[@xxx] | 获得签到背景\n- \n- 查看等级排名\n- 查看我的钱包\n- ",
	}
	engine = zero.NewTemplate(metaData)
	styles = []scoredrawer{
		drawScore15,
		drawScore16,
		drawScore17,
		drawScore17b2,
	}
)

func init() {
	cachePath = filepath.Join(path.GetDataPath(), "cache")
	sdb = initialize()
	file.CreateIfNotExist(cachePath)

	engine.OnRegex(`^签到\s?(\d*)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 选择key
		key := ctx.State["regex_matched"].([]string)[1]
		k := uint8(0)
		if key != "" {
			kn, err := strconv.Atoi(key)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			k = uint8(kn)
		}
		if int(k) >= len(styles) {
			ctx.SendChain(message.Text("ERROR: 未找到签到设定: ", key))
			return
		}
		uid := transform.BidWithuidInt64(ctx)
		today := time.Now().Format("20060102")
		// 签到图片
		drawedFile := strconv.FormatInt(uid, 10) + today + "signin.png"
		picFile := filepath.Join(cachePath, strconv.FormatInt(uid, 10)+today)
		// 获取签到时间
		si := sdb.GetSignInByUID(uid)
		siUpdateTimeStr := si.UpdatedAt.Format("20060102")
		switch {
		case si.Count >= signinMax && siUpdateTimeStr == today:
			// 如果签到时间是今天
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("今天你已经签到过了！"))
			res, _ := cache.GetRedisCli().Get(context.TODO(), drawedFile).Result()
			if res != "" {
				ctx.SendChain(message.ImageBytes([]byte(res)))
			}
			return
		case siUpdateTimeStr != today:
			// 如果是跨天签到就清数据
			err := sdb.ResetTable()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		}
		// 更新签到次数
		err := sdb.InsertOrUpdateSignInCountByUID(uid, si.Count+1)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 更新经验
		level := sdb.GetScoreByUID(uid).Score + 1
		if level > SCOREMAX {
			level = SCOREMAX
			ctx.SendChain(message.At(uid), message.Text("你的等级已经达到上限"))
		}
		err = sdb.InsertOrUpdateScoreByUID(uid, level)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 更新钱包
		rank := getrank(level)
		add := 1 + rand.IntN(10) + rank*5 // 等级越高获得的钱越高
		go func() {
			err = wallet.UpdateWalletByCtx(ctx, add)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		}()
		alldata := &scdata{
			drawedfile: drawedFile,
			picfile:    picFile,
			uid:        ctx.Event.UserID,
			nickname:   ctx.CardOrNickName(ctx.Event.UserID),
			inc:        add,
			score:      wallet.GetWalletMoneyByCtx(ctx),
			level:      level,
			rank:       rank,
		}
		drawimage, err := styles[k](alldata)
		if err != nil {
			ctx.SendError(err)
			return
		}
		// done.
		data, err := imgfactory.ToBytes(drawimage)
		_, err = cache.GetRedisCli().Set(context.TODO(), drawedFile, data, 24*time.Hour).Result()
		if err != nil {
			ctx.SendError(err)
			return
		}
		ctx.SendChain(message.ImageBytes(data))
		// trySendImage(drawedFile, ctx)
	})

	engine.OnPrefix("获得签到背景", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			atInfos := ctx.GetAtInfos()
			uid := ctx.Event.UserID
			if len(atInfos) == 1 {
				uid = atInfos[0].QQ
			}
			uid = transform.BidWithidInt64Custom(ctx.Event.SelfID, uid)
			picFile := filepath.Join(cachePath, strconv.FormatInt(uid, 10)+time.Now().Format("20060102")+".png")
			if file.IsNotExist(picFile) {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请先签到！"))
				return
			}
			trySendImage(picFile, ctx)
		})
	engine.OnFullMatch("查看等级排名", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			today := time.Now().Format("20060102")
			drawedFile := cachePath + today + "scoreRank.png"
			if file.IsExist(drawedFile) {
				trySendImage(drawedFile, ctx)
				return
			}
			st, err := sdb.GetScoreRankByTopN(10)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if len(st) == 0 {
				ctx.SendChain(message.Text("ERROR: 目前还没有人签到过"))
				return
			}
			_, err = os.ReadFile(text.FontFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			b, err := os.ReadFile(text.FontFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			font, err := freetype.ParseFont(b)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			f, err := os.Create(drawedFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			var bars []chart.Value
			for _, v := range st {
				if v.Score != 0 {
					bars = append(bars, chart.Value{
						Label: ctx.CardOrNickName(v.UID),
						Value: float64(v.Score),
					})
				}
			}
			err = chart.BarChart{
				Font:  font,
				Title: "等级排名(1天只刷新1次)",
				Background: chart.Style{
					Padding: chart.Box{
						Top: 40,
					},
				},
				YAxis: chart.YAxis{
					Range: &chart.ContinuousRange{
						Min: 0,
						Max: math.Ceil(bars[0].Value/10) * 10,
					},
				},
				Height:   500,
				BarWidth: 50,
				Bars:     bars,
			}.Render(chart.PNG, f)
			_ = f.Close()
			if err != nil {
				_ = os.Remove(drawedFile)
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			trySendImage(drawedFile, ctx)
		})
	engine.OnRegex(`^设置签到预设\s*(\d+)$`, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		key := ctx.State["regex_matched"].([]string)[1]
		kn, err := strconv.Atoi(key)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		k := uint8(kn)
		if int(k) >= len(styles) {
			ctx.SendChain(message.Text("ERROR: 未找到签到设定: ", key))
			return
		}
		// gid := ctx.Event.GroupID
		// if gid == 0 {
		// 	gid = -ctx.Event.UserID
		// }
		// err = ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).SetData(gid, int64(k))
		// if err != nil {
		// 	ctx.SendChain(message.Text("ERROR: ", err))
		// 	return
		// }
		ctx.SendChain(message.Text("设置成功"))
	})
}

func getHourWord(t time.Time) string {
	h := t.Hour()
	switch {
	case 6 <= h && h < 12:
		return "早上好"
	case 12 <= h && h < 14:
		return "中午好"
	case 14 <= h && h < 19:
		return "下午好"
	case 19 <= h && h < 24:
		return "晚上好"
	case 0 <= h && h < 6:
		return "凌晨好"
	default:
		return ""
	}
}

func getrank(count int) int {
	for k, v := range rankArray {
		if count == v {
			return k
		} else if count < v {
			return k - 1
		}
	}
	return -1
}

func initPic(picFile string, uid int64) (avatar []byte, err error) {
	var response *http.Response
	if uid != 0 {
		avatar, err = cache.GetAvatar(uid)
		if err != nil {
			return
		}
	}
	if file.IsExist(picFile) {
		return
	}
	url, err := net_tools.GetRealURL(backgroundURL)

	if err == nil {
		request, _ := http.NewRequest("GET", url, nil)
		request.Header.Add("referer", referer)
		response, err = http.DefaultClient.Do(request)

		if response.StatusCode != http.StatusOK {
			s := fmt.Sprintf("status code: %d", response.StatusCode)
			err = errors.New(s)
		}
	}

	var data []byte
	if err != nil {
		var files []fs.DirEntry
		files, err = os.ReadDir(cachePath)
		if err != nil {
			return
		}
		if len(files) <= 0 {
			err = errors.New("no files")
			return
		}
		randomIndex := rand.IntN(len(files))
		selectedFile := files[randomIndex]

		data, err = os.ReadFile(filepath.Join(cachePath, selectedFile.Name()))
	} else {
		data, err = io.ReadAll(response.Body)
		response.Body.Close()
		// 如果图片是webp格式，则转换为png
		if err == nil {
			img, err := webp.Decode(bytes.NewReader(data))
			if err == nil {
				return data, err
			}
			buf := bytes.NewBuffer(nil)
			err = png.Encode(buf, img)
			if err == nil {
				data = buf.Bytes()
			}
		}
	}

	if err != nil {
		return
	}

	return avatar, os.WriteFile(picFile, data, 0644)
}

// 使用"file:"发送图片失败后，改用base64发送
func trySendImage(filePath string, ctx *zero.Ctx) {
	if _, err := ctx.SendChain(message.Image("file:///" + filePath)); err == nil {
		return
	}
	imgFile, err := os.Open(filePath)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: 无法打开文件", err))
		return
	}
	defer imgFile.Close()
	// 使用 base64.NewEncoder 将文件内容编码为 base64 字符串
	var encodedFileData strings.Builder
	encodedFileData.WriteString("base64://")
	encoder := base64.NewEncoder(base64.StdEncoding, &encodedFileData)
	_, err = io.Copy(encoder, imgFile)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: 无法编码文件内容", err))
		return
	}
	encoder.Close()
	drawedFileBase64 := encodedFileData.String()
	if _, err := ctx.SendChain(message.Image(drawedFileBase64)); err != nil {
		ctx.SendChain(message.Text("ERROR: 无法读取图片文件", err))
		return
	}
}
