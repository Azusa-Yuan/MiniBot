package service

import (
	"MiniBot/utils/resource"
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
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

	allocCtx, cancel1 = chromedp.NewExecAllocator(context.Background(), opts...)
	Ctx, cancel2      = chromedp.NewContext(allocCtx)
)

func init() {
	resource.ResourceManager.Register(
		func(ctx context.Context) error {
			cancel2()
			cancel1()
			return nil
		},
	)
}

func Screenshot(server string, productType string) ([]byte, string, error) {

	if _, ok := ReportRegions[server]; !ok {
		return nil, "", nil
	}

	// 导航到指定的URL
	var buf []byte
	url := fmt.Sprintf("https://www.yxdr.com/bijiaqi/dnf/%s/kua%s", productType, ReportRegions[server])
	err := chromedp.Run(Ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("#right_m"),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return nil, url, err
	}

	return buf, url, nil
}
