package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
)

type demo struct {
	metrics *prometheus.GaugeVec
	arg     string
}

func (e *demo)Describe(desc chan<- *prometheus.Desc)  {
	e.metrics.Describe(desc)
}

func (e *demo)Collect(metrics chan<- prometheus.Metric)  {
	label := map[string]string{"demo": e.arg}
	e.metrics.With(label).Set(float64(rand.Int()))
	e.metrics.Collect(metrics)
}

func NewDemo(arg string) *demo {
	e := demo{
		metrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "test1",
				Help: "This is test1",
			},
			[]string{"demo"},
		) ,
		arg: arg,
	}
	return &e
}
