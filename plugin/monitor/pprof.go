package monitor

import (
	"MiniBot/service/web"

	"github.com/gin-contrib/pprof"
)

func init() {
	r := web.GetWebEngine()
	pprof.Register(r)
}
