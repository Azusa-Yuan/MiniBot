package net_tools

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

var (
	defaultClient          = &http.Client{}
	clientWithoutTLSVerify = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 禁用证书验证（仅用于测试）
			},
		},
	}
)

// GetRealURL 获取跳转后的链接
func GetRealURL(url string) (realurl string, err error) {
	data, err := http.Head(url)
	if err != nil {
		return
	}
	_ = data.Body.Close()
	realurl = data.Request.URL.String()
	return
}

func download(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status, ":", string(data))
	}
	return data, nil
}

func Download(url string) ([]byte, error) {
	return download(defaultClient, url)
}

func DownloadWithoutTLSVerify(url string) ([]byte, error) {
	return download(clientWithoutTLSVerify, url)
}
