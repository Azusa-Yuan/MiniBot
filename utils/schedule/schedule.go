package schedule

import (
	"time"

	"github.com/fumiama/cron"
)

var (
	timeZone, _ = time.LoadLocation("Asia/Shanghai")
	// 不建议频繁进行的cron放在这里
	Cron = cron.New(cron.WithLocation(timeZone))
)
