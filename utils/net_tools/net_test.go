package net_tools

import (
	"fmt"
	"testing"
)

func TestXxx(t *testing.T) {
	_, err := DownloadWithoutTLSVerify("https://multimedia.nt.qq.com.cn/download?appid=1406&fileid=CgoxMDQzNzI4NDE3EhS-gq3plBoE1JXQl8si_qjkrWd0Ghjhvj0g_gootuDmxMOwiAM&spec=0&rkey=CAESKBkcro_MGujoxQ0McQzXL6ZGy9Dc9MaKX_qMjdQu4zlOGf_sOdxCkV8")
	fmt.Println(err)
}
