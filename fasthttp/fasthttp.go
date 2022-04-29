package fasthttp

import (
	"fmt"
	"github.com/gohutool/log4go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : prometheus.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/29 12:16
* 修改历史 : 1. [2022/4/29 12:16] 创建文件 by LongYong
*/

var fasthttp_logger = log4go.LoggerManager.GetLogger("gohutool.prometheus4go.fasthttp")

const (
	hdrAccept             = "Accept"
	contentTypeHeader     = "Content-Type"
	contentEncodingHeader = "Content-Encoding"
	acceptEncodingHeader  = "Accept-Encoding"

	REQUEST_OTHER = "Other"
	REQUEST_ALL   = "All"
)

func ParseFormat(acceptHeader string, enableOpenMetrics bool) expfmt.Format {
	header := make(http.Header)
	header.Set(hdrAccept, acceptHeader)
	var contentType expfmt.Format
	if enableOpenMetrics {
		contentType = expfmt.NegotiateIncludingOpenMetrics(header)
	} else {
		contentType = expfmt.Negotiate(header)
	}
	return contentType
}

var Request_Metrics_Codes = map[string]int{
	strconv.Itoa(http.StatusOK):                      1,
	strconv.Itoa(http.StatusFound):                   1,
	strconv.Itoa(http.StatusNotModified):             1,
	strconv.Itoa(http.StatusUseProxy):                1,
	strconv.Itoa(http.StatusSeeOther):                1,
	strconv.Itoa(http.StatusTemporaryRedirect):       1,
	strconv.Itoa(http.StatusPermanentRedirect):       1,
	strconv.Itoa(http.StatusBadRequest):              1,
	strconv.Itoa(http.StatusUnauthorized):            1,
	strconv.Itoa(http.StatusMethodNotAllowed):        1,
	strconv.Itoa(http.StatusForbidden):               1,
	strconv.Itoa(http.StatusNotFound):                1,
	strconv.Itoa(http.StatusInternalServerError):     1,
	strconv.Itoa(http.StatusNotImplemented):          1,
	strconv.Itoa(http.StatusBadGateway):              1,
	strconv.Itoa(http.StatusServiceUnavailable):      1,
	strconv.Itoa(http.StatusGatewayTimeout):          1,
	strconv.Itoa(http.StatusHTTPVersionNotSupported): 1,
}

func init() {
	cnt := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "promhttp_metric_handler_requests_total",
			Help: "Total number of scrapes by HTTP status code.",
		},
		[]string{"code"},
	)
	// Initialize the most likely HTTP status codes.
	for k, _ := range Request_Metrics_Codes {
		cnt.WithLabelValues(k)
	}

	cnt.WithLabelValues(REQUEST_OTHER)

	//cnt.WithLabelValues("200")
	//cnt.WithLabelValues("500")
	//cnt.WithLabelValues("503")
	//cnt.WithLabelValues(http.StatusNotFound)

	if err := prometheus.DefaultRegisterer.Register(cnt); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			cnt = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			fasthttp_logger.Error("%v", err)
		}
	}

	requestCounter = cnt
}

func RequestCounterHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		if next != nil {
			next(ctx)
		}

		func() {
			defer func() {
				if err := recover(); err != nil {
					fasthttp_logger.Error("%v", err)
				}
			}()
			code := strconv.Itoa(ctx.Response.StatusCode())
			if _, ok := Request_Metrics_Codes[code]; ok {
				requestCounter.WithLabelValues(code).Inc()
			} else {
				requestCounter.WithLabelValues(REQUEST_OTHER).Inc()
			}

			requestCounter.WithLabelValues(REQUEST_ALL).Inc()

		}()
	})
}

func PrometheusHandlerFor(opts HandlerOpts, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		next(ctx)
		PrometheusHandler(opts)(ctx)
	})
}

func PrometheusHandler(opts HandlerOpts) fasthttp.RequestHandler {
	return InstrumentMetricHandler(prometheus.DefaultRegisterer,
		HandlerFor(prometheus.DefaultGatherer, opts))
}

var requestCounter *prometheus.CounterVec

func InstrumentMetricHandler(reg prometheus.Registerer, handler fasthttp.RequestHandler) fasthttp.RequestHandler {

	gge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "promhttp_metric_handler_requests_in_flight",
		Help: "Current number of scrapes being served.",
	})
	if err := reg.Register(gge); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			gge = are.ExistingCollector.(prometheus.Gauge)
		} else {
			panic(err)
		}
	}

	return InstrumentHandlerInFlight(gge, handler)
}

func InstrumentHandlerInFlight(g prometheus.Gauge, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		g.Inc()
		defer g.Dec()
		next(ctx)
	})
}

func HandlerFor(reg prometheus.Gatherer, opts HandlerOpts) fasthttp.RequestHandler {

	var (
		inFlightSem chan struct{}
		errCnt      = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "promhttp_metric_handler_errors_total",
				Help: "Total number of internal errors encountered by the promhttp metric handler.",
			},
			[]string{"cause"},
		)
	)

	if opts.MaxRequestsInFlight > 0 {
		inFlightSem = make(chan struct{}, opts.MaxRequestsInFlight)
	}
	if opts.Registry != nil {
		// Initialize all possibilities that can occur below.
		errCnt.WithLabelValues("gathering")
		errCnt.WithLabelValues("encoding")
		if err := opts.Registry.Register(errCnt); err != nil {
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
				errCnt = are.ExistingCollector.(*prometheus.CounterVec)
			} else {
				panic(err)
			}
		}
	}

	h := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		if inFlightSem != nil {
			select {
			case inFlightSem <- struct{}{}: // All good, carry on.
				defer func() { <-inFlightSem }()
			default:
				ctx.Error(fmt.Sprintf(
					"Limit of concurrent requests reached (%d), try again later.", opts.MaxRequestsInFlight,
				), http.StatusServiceUnavailable)

				return
			}
		}

		mfs, err := reg.Gather()

		if err != nil {
			if opts.ErrorLog != nil {
				opts.ErrorLog.Println("error gathering metrics:", err)
			}
			errCnt.WithLabelValues("gathering").Inc()
			switch opts.ErrorHandling {
			case PanicOnError:
				panic(err)
			case ContinueOnError:
				if len(mfs) == 0 {
					// Still report the error if no metrics have been gathered.
					httpError(ctx, err)
					return
				}
			case HTTPErrorOnError:
				httpError(ctx, err)
				return
			}
		}

		var contentType expfmt.Format
		contentType = ParseFormat(string(ctx.Request.Header.Peek(hdrAccept)), opts.EnableOpenMetrics)
		ctx.Response.Header.Set(contentTypeHeader, string(contentType))

		enc := expfmt.NewEncoder(ctx.Response.BodyWriter(), contentType)

		// handleError handles the error according to opts.ErrorHandling
		// and returns true if we have to abort after the handling.
		handleError := func(err error) bool {
			if err == nil {
				return false
			}
			if opts.ErrorLog != nil {
				opts.ErrorLog.Println("error encoding and sending metric family:", err)
			}
			errCnt.WithLabelValues("encoding").Inc()
			switch opts.ErrorHandling {
			case PanicOnError:
				panic(err)
			case HTTPErrorOnError:
				// We cannot really send an HTTP error at this
				// point because we most likely have written
				// something to rsp already. But at least we can
				// stop sending.
				return true
			}
			// Do nothing in all other cases, including ContinueOnError.
			return false
		}

		for _, mf := range mfs {
			if handleError(enc.Encode(mf)) {
				return
			}
		}

		if closer, ok := enc.(expfmt.Closer); ok {
			// This in particular takes care of the final "# EOF\n" line for OpenMetrics.
			if handleError(closer.Close()) {
				return
			}
		}
	})

	if opts.Timeout <= 0 {
		return h
	}

	return fasthttp.TimeoutHandler(h, opts.Timeout, fmt.Sprintf(
		"Exceeded configured timeout of %v.\n",
		opts.Timeout),
	)
}

// httpError removes any content-encoding header and then calls http.Error with
// the provided error and http.StatusInternalServerError. Error contents is
// supposed to be uncompressed plain text. Same as with a plain http.Error, this
// must not be called if the header or any payload has already been sent.
func httpError(ctx *fasthttp.RequestCtx, err error) {
	ctx.Response.Header.Del(contentEncodingHeader)
	ctx.Error(
		"An error has occurred while serving metrics:\n\n"+err.Error(),
		http.StatusInternalServerError,
	)
}
