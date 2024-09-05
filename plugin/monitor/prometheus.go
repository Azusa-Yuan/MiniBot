package monitor

import (
	"MiniBot/service/web"
	zero "ZeroBot"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_time_seconds",
			Help:    "Histogram of processing time for consumers.",
			Buckets: []float64{0.5, 1.0, 2.0, 5.0, 10.0}, // 自定义桶
		},
		[]string{"plugin", "controller", "success"},
	)

	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"plugin", "controller", "success"},
	)
)

func init() {
	// 注册指标
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(RequestsTotal)

	r := web.GetWebEngine()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	zero.GolbaleMiddleware.Before(
		func(ctx *zero.Ctx) {
			ctx.State["time"] = time.Now()
		},
	)
	zero.GolbaleMiddleware.After(
		func(ctx *zero.Ctx) {
			startTime := ctx.State["time"].(time.Time)
			status := "ok"
			if ctx.Err != nil {
				status = "err"
				ctx.Err = nil
			}
			pluginName := ctx.GetMatcherMetadata().PluginName
			if pluginName == "" {
				pluginName = "default"
			}
			RequestsTotal.WithLabelValues(pluginName, "", status).Inc()
			ResponseTime.WithLabelValues(pluginName, "", status).Observe(float64(time.Since(startTime).Microseconds()) / 1000)
		},
	)
}
