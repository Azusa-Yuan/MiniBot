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
			Help:    "Histogram of processing time",
			Buckets: []float64{0.5, 1.0, 2.0, 5.0, 10.0}, // 自定义桶
		},
		[]string{"plugin", "controller", "success"},
	)

	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of requests",
		},
		[]string{"plugin", "matcher", "success"},
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
			matcherMetadata := ctx.GetMatcherMetadata()
			RequestsTotal.WithLabelValues(matcherMetadata.PluginName, matcherMetadata.MatcherName, status).Inc()
			ResponseTime.WithLabelValues(matcherMetadata.PluginName, matcherMetadata.MatcherName, status).Observe(float64(time.Since(startTime).Milliseconds()) / 1000)
		},
	)
}
