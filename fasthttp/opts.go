package fasthttp

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : opts.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/29 12:27
* 修改历史 : 1. [2022/4/29 12:27] 创建文件 by LongYong
*/

// HandlerErrorHandling defines how a Handler serving metrics will handle
// errors.
type HandlerErrorHandling int

// These constants cause handlers serving metrics to behave as described if
// errors are encountered.
const (

	// Serve an HTTP status code 500 upon the first error
	// encountered. Report the error message in the body. Note that HTTP
	// errors cannot be served anymore once the beginning of a regular
	// payload has been sent. Thus, in the (unlikely) case that encoding the
	// payload into the negotiated wire format fails, serving the response
	// will simply be aborted. Set an ErrorLog in HandlerOpts to detect
	// those errors.

	HTTPErrorOnError HandlerErrorHandling = iota

	// Ignore errors and try to serve as many metrics as possible.  However,
	// if no metrics can be served, serve an HTTP status code 500 and the
	// last error message in the body. Only use this in deliberate "best
	// effort" metrics collection scenarios. In this case, it is highly
	// recommended providing other means of detecting errors: By setting an
	// ErrorLog in HandlerOpts, the errors are logged. By providing a
	// Registry in HandlerOpts, the exposed metrics include an error counter
	// "prompt_metric_handler_errors_total", which can be used for
	// alerts.

	ContinueOnError

	// PanicOnError Panic upon the first error encountered (useful for "crash only" apps).

	PanicOnError
)

// Logger is the minimal interface HandlerOpts needs for logging. Note that
// log.Logger from the standard library implements this interface, and it is
// easy to implement by custom loggers, if they don't do so already anyway.
type Logger interface {
	Println(v ...interface{})
}

// HandlerOpts specifies options how to serve metrics via an http.Handler. The
// zero value of HandlerOpts is a reasonable default.
type HandlerOpts struct {
	// ErrorLog specifies an optional Logger for errors collecting and
	// serving metrics. If nil, errors are not logged at all. Note that the
	// type of a reported error is often prometheus.MultiError, which
	// formats into a multi-line error string. If you want to avoid the
	// latter, create a Logger implementation that detects a
	// prometheus.MultiError and formats the contained errors into one line.
	ErrorLog Logger
	// ErrorHandling defines how errors are handled. Note that errors are
	// logged regardless of the configured ErrorHandling provided ErrorLog
	// is not nil.
	ErrorHandling HandlerErrorHandling
	// If Registry is not nil, it is used to register a metric
	// "promhttp_metric_handler_errors_total", partitioned by "cause". A
	// failed registration causes a panic. Note that this error counter is
	// different from the instrumentation you get from the various
	// InstrumentHandler... helpers. It counts errors that don't necessarily
	// result in a non-2xx HTTP status code. There are two typical cases:
	// (1) Encoding errors that only happen after streaming of the HTTP body
	// has already started (and the status code 200 has been sent). This
	// should only happen with custom collectors. (2) Collection errors with
	// no effect on the HTTP status code because ErrorHandling is set to
	// ContinueOnError.
	Registry prometheus.Registerer
	// If DisableCompression is true, the handler will never compress the
	// response, even if requested by the client.
	DisableCompression bool
	// The number of concurrent HTTP requests is limited to
	// MaxRequestsInFlight. Additional requests are responded to with 503
	// Service Unavailable and a suitable message in the body. If
	// MaxRequestsInFlight is 0 or negative, no limit is applied.
	MaxRequestsInFlight int
	// If handling a request takes longer than Timeout, it is responded to
	// with 503 ServiceUnavailable and a suitable Message. No timeout is
	// applied if Timeout is 0 or negative. Note that with the current
	// implementation, reaching the timeout simply ends the HTTP requests as
	// described above (and even that only if sending of the body hasn't
	// started yet), while the bulk work of gathering all the metrics keeps
	// running in the background (with the eventual result to be thrown
	// away). Until the implementation is improved, it is recommended to
	// implement a separate timeout in potentially slow Collectors.
	Timeout time.Duration
	// If true, the experimental OpenMetrics encoding is added to the
	// possible options during content negotiation. Note that Prometheus
	// 2.5.0+ will negotiate OpenMetrics as first priority. OpenMetrics is
	// the only way to transmit exemplars. However, the move to OpenMetrics
	// is not completely transparent. Most notably, the values of "quantile"
	// labels of Summaries and "le" labels of Histograms are formatted with
	// a trailing ".0" if they would otherwise look like integer numbers
	// (which changes the identity of the resulting series on the Prometheus
	// server).
	EnableOpenMetrics bool
}
