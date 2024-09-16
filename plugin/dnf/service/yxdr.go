package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
)

var (
	ReportRegions = map[string]string{
		"跨1":  "1",
		"跨2":  "2",
		"跨3a": "3a",
		"跨3b": "3b",
		"跨4":  "4",
		"跨5":  "5",
		"跨6":  "6",
		"跨7":  "7",
		"跨8":  "8",
		"跨一":  "1",
		"跨二":  "2",
		"跨三A": "3a",
		"跨三B": "3b",
		"跨四":  "4",
		"跨五":  "5",
		"跨六":  "6",
		"跨七":  "7",
		"跨八":  "8",
	}

	// 设置浏览器选项
	opts = append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("remote-debugging-port", "9222"),
	)

	allocCtx, _ = chromedp.NewExecAllocator(context.Background(), opts...)
	TemCtx, _   = chromedp.NewContext(allocCtx)
	ScCtx       = context.WithoutCancel(TemCtx)
)

func init() {
	chromedp.Emulate(device.Reset)
	chromedp.EmulateViewport(1000, 1500)
}

func Screenshot(server string, productType string) ([]byte, string, error) {

	if _, ok := ReportRegions[server]; !ok {
		return nil, "", nil
	}

	// 导航到指定的URL
	var buf []byte
	ctx, _ := context.WithTimeout(ScCtx, 20*time.Second)
	// url := "https://www.baidu.com"
	url := fmt.Sprintf("https://www.yxdr.com/bijiaqi/dnf/%s/kua%s", productType, ReportRegions[server])
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#right_m"),
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		return nil, url, err
	}

	return buf, url, nil
}
