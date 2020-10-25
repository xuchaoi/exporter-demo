package main

import (
	"awesomeProject/pkg/metrics"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	// exporter默认端口：8080
	addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	// promauto.NewCounter方法会自动将新建的Counter（只增不减）指标注册到metrics中
	exampleCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "example_count_total",
		Help: "The total number of exporter-demo",
	})
	// promauto.NewGauge方法会自动将新建的Gauge（可增可减）指标注册到metrics中
	exampleGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "example_gauge_value",
		Help: "The current number of exporter-demo",
		// 在标签中指定label
		ConstLabels: map[string]string{"example": "test"},
	})
	// prometheus.NewSummaryVec生成一个自定义Label的Summary（分位图：各个分位，采样点sum、count累积和）指标，但未注册
	exampleSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "example_summary_seconds",
			Help: "The latency distributions Summary of exporter-demo",
			// 5分位（误差不超过0.05，0.45~0.55），9分位（误差不超过0.01，0.89~0.91），99分位（误差不超过0.001，0.989~0.991）
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"service"},
	)
	// prometheus.NewHistogram生成一个自定义Label的Histogram（柱状图：各个区间，采样点sum、count累积和）指标，但未注册
	exampleHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "example_histogram_seconds",
		Help: "The latency distributions histogram of exporter-demo",
		// 区间从0开始，每隔5个长度为一个区间，一共20个区间，每个区间也叫桶
		Buckets: prometheus.LinearBuckets(0, 5, 20),
	})
)

func init()  {
	// 将exampleSummary/exampleHistogram指标注册到metrics中
	prometheus.MustRegister(exampleSummary)
	prometheus.MustRegister(exampleHistogram)
	// 注册自定义的metrics
	demo := metrics.NewDemo("hello")
	prometheus.MustRegister(demo)
	// 将程序构建相关信息放到metrics中
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

func main() {
	// 获取arg参数，更新启动参数
	flag.Parse()

	// 程序定时更新exampleCount/exampleGauge的metrics值
	go func() {
		for {
			exampleCount.Inc()
			exampleGauge.Add(10)
			time.Sleep(5 * time.Second)
		}
	}()
	// 程序定时更新summary、histogram指标
	go func() {
		for {
			v := rand.Float64()
			exampleSummary.WithLabelValues("uniform").Observe(v)
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		for {
			v := rand.NormFloat64()
			exampleSummary.WithLabelValues("normal").Observe(v)
			exampleHistogram.(prometheus.ExemplarObserver).ObserveWithExemplar(
				// 为指标赋值时，添加label
				v, prometheus.Labels{"dummyID": fmt.Sprint(rand.Intn(100000))},
			)
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		for {
			v := rand.ExpFloat64() / 1e6
			exampleSummary.WithLabelValues("exponential").Observe(v)
			time.Sleep(5 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
