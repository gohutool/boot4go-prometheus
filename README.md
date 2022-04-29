# boot4go-prometheus
a prometheus exporter support for fasthttp


![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

# Introduce
This project is depend on fasthttp, prometheus and log4go , It is a prometheus export client support for fasthttp

# Feature
- Prometheus export client in fasthttp
- A Request status metrics in http application

# Usage
- Add boot4go-prometheus with the following import

```
import (
prometheusfasthttp "github.com/gohutool/boot4go-prometheus/fasthttp"
)
```

- Add PrometheusHandler and map the metrics path 

```
handler := func(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/metrics":
		prometheusfasthttp.PrometheusHandler(prometheusfasthttp.HandlerOpts{})(ctx)
	case "/sample1":
		sample1HandlerFunc(ctx)
	case "/sample2":
		sample2HandlerFunc(ctx)
	default:
		ctx.Error("not found", fasthttp.StatusNotFound)
	}
}
fasthttp.ListenAndServe(":80", handler)
```

or

```
handler := prometheusfasthttp.PrometheusHandlerFor(prometheusfasthttp.HandlerOpts{}, func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
            case "/metrics":
                return
            case "/sample1":
                sample1HandlerFunc(ctx)
            case "/sample2":
                sample2HandlerFunc(ctx)
            default:
                ctx.Error("not found", fasthttp.StatusNotFound)
            }
	})
fasthttp.ListenAndServe(":80", handler)
```

- If you have use thirdparty router, such as fasthttp-router, you can add the PrometheusHandler as below

```
func InitRouter(router *routing.Router) {
	router.Get("/metrics", func(context *routing.Context) error {
		prometheusfasthttp.PrometheusHandler(prometheusfasthttp.HandlerOpts{})(context.RequestCtx)
		return nil
	})
}
```

- How to add some metrics by customizer

```
    prometheus.MustRegister(totalCounterVec)
	prometheus.MustRegister(amountSummaryVec)
	prometheus.MustRegister(amountGaugeVec)
```

- How to add default ResponseStatus metrics in your fasthttp application

```
	requestHandler := func(ctx *fasthttp.RequestCtx) {

		Logger.Debug("%v %v %v %v", string(ctx.Path()), ctx.URI().String(), string(ctx.Method()), ctx.QueryArgs().String())
		defer func() {
			if err := recover(); err != nil {
				Logger.Debug(err)
				// ctx.Error(fmt.Sprintf("%v", err), http.StatusInternalServerError)
				Error(ctx, Result.Fail(fmt.Sprintf("%v", err)).Json(), http.StatusInternalServerError)
			}

			ctx.Response.Header.Set("tick", time.Now().String())
			ctx.Response.Header.SetServer("Gateway-UIManager")

			prometheusfasthttp.RequestCounterHandler(nil)(ctx)

			Logger.Debug("router.HandleRequest is finish")

		}()

		router.HandleRequest(ctx)
	}
	
	
	// Start HTTP server.
	Logger.Info("Starting HTTP server on %v", listener.Addr().String())
	go func() {
		if err := fasthttp.Serve(listener, requestHandler); err != nil {
			Logger.Critical("error in ListenAndServe: %v", err)
		}
	}()
```

- visit http://localhost/metrics

```
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 7
promhttp_metric_handler_requests_total{code="302"} 0
promhttp_metric_handler_requests_total{code="303"} 0
promhttp_metric_handler_requests_total{code="304"} 0
promhttp_metric_handler_requests_total{code="305"} 0
promhttp_metric_handler_requests_total{code="307"} 0
promhttp_metric_handler_requests_total{code="308"} 0
promhttp_metric_handler_requests_total{code="400"} 0
promhttp_metric_handler_requests_total{code="401"} 0
promhttp_metric_handler_requests_total{code="403"} 0
promhttp_metric_handler_requests_total{code="404"} 0
promhttp_metric_handler_requests_total{code="405"} 0
promhttp_metric_handler_requests_total{code="500"} 2
promhttp_metric_handler_requests_total{code="501"} 0
promhttp_metric_handler_requests_total{code="502"} 0
promhttp_metric_handler_requests_total{code="503"} 0
promhttp_metric_handler_requests_total{code="504"} 0
promhttp_metric_handler_requests_total{code="505"} 0
promhttp_metric_handler_requests_total{code="All"} 9
promhttp_metric_handler_requests_total{code="Other"} 0
# HELP worker_jobs_processed_gauge
# TYPE worker_jobs_processed_gauge gauge
worker_jobs_processed_gauge{BeanName="Hello"} 0
```

# Related Project

- Fasthttp  https://github.com/valyala/fasthttp  Fast HTTP package for Go.
- Prometheus https://github.com/prometheus/client_golang  Prometheus instrumentation library for Go applications
- Log4go https://github.com/gohutool/log4go A logkit like as log4j with go language
