package monitor

import (
	"MiniBot/service/web"

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

}
