package fasthttp

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"os"
	"testing"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : prometheus_test.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/29 09:18
* 修改历史 : 1. [2022/4/29 09:18] 创建文件 by LongYong
*/

var (
	totalCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "worker",
			Subsystem: "jobs",
			Name:      "processed_total",
			Help:      "Total number of jobs processed by the workers",
		},
		// We will want to monitor the worker ID that processed the
		// job, and the type of job that was processed
		[]string{"worker_id", "type"},
	)
	//totalCounterVec.WithLabelValues("num", "counter").Inc()
	/*
	   3 times
	   # HELP worker_jobs_processed_total Total number of jobs processed by the workers
	   # TYPE worker_jobs_processed_total counter
	   worker_jobs_processed_total{type="counter",worker_id="num"} 3
	*/

	amountGaugeVec = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   "worker",
			Subsystem:   "jobs",
			Name:        "processed_gauge",
			ConstLabels: map[string]string{"BeanName": "Hello"},
		},
	)
	//amountGaugeVec.Set(10)
	/*
	   3 times
	   # HELP worker_jobs_processed_gauge
	   # TYPE worker_jobs_processed_gauge gauge
	   worker_jobs_processed_gauge{BeanName="Hello"} 10
	*/

	amountSummaryVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "worker",
			Subsystem: "jobs",
			Name:      "processed_summary",
			Help:      "Total number of jobs processed by the workers",
		}, []string{"worker_id", "type"},
	)
	//amountSummaryVec.WithLabelValues("num", "counter").Observe(11)
	/*
	   3 times
	   # HELP worker_jobs_processed_summary Total number of jobs processed by the workers
	   # TYPE worker_jobs_processed_summary summary
	   worker_jobs_processed_summary_sum{type="counter",worker_id="num"} 33
	   worker_jobs_processed_summary_count{type="counter",worker_id="num"} 3
	*/
)

func TestPrometheus(t *testing.T) {

	prometheus.MustRegister(totalCounterVec)
	prometheus.MustRegister(amountSummaryVec)
	prometheus.MustRegister(amountGaugeVec)

	totalCounterVec.WithLabelValues("num", "counter").Inc()
	totalCounterVec.WithLabelValues("num", "counter").Inc()
	totalCounterVec.WithLabelValues("num", "counter").Inc()
	amountGaugeVec.Set(10)
	amountSummaryVec.WithLabelValues("num", "counter").Observe(11)

	m := &dto.Metric{}
	fmt.Println(m.String())
	fmt.Println(totalCounterVec.WithLabelValues("num", "counter").Desc().String())

	if mfs, err := prometheus.DefaultGatherer.Gather(); err == nil {
		fmt.Println(len(mfs))

		var contentType expfmt.Format
		contentType = ParseFormat("", false)
		enc := expfmt.NewEncoder(os.Stdout, contentType)

		for _, mf := range mfs {
			enc.Encode(mf)
		}

		if err := enc.(expfmt.Closer).Close(); err == nil {
			fmt.Println("OK")
		} else {
			fmt.Println("Error ", err.Error())
		}

	} else {

		fmt.Println("Error ", err.Error())
	}

}
