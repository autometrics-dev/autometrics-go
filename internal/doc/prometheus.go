package doc

import (
	"fmt"
	"net/url"
)

// TODO: This should be read from an argument of go generate
const prometheusInstanceUrl = "http://localhost:9090/"

// TODO: Use globals from somewhere else in the lib
const (
	callCounterMetricName     = "function_calls_counter"
	callDurationMetricName    = "function_calls_duration_bucket"
	concurrentCallsMetricName = "function_calls_concurrent"
)

// https://github.com/autometrics-dev/autometrics-rs/blob/0d9c2118e7d9f77032a12e0a2fadb15d9c28446e/autometrics-macros/src/lib.rs#L314

type Prometheus struct {
	instanceUrl url.URL
}

func NewPrometheusDoc() Prometheus {
	// No way to have a url.URL constant, so we reparse it here

	prometheusInstanceUrl, _ := url.Parse(prometheusInstanceUrl)
	return Prometheus{instanceUrl: *prometheusInstanceUrl}
}

func (p Prometheus) makePrometheusUrl(query, comment string) url.URL {
	ret := p.instanceUrl

	params := ret.Query()
	commentAndQuery := fmt.Sprintf("# %s\n\n%s", comment, query)
	params.Add("g0.expr", commentAndQuery)
	// Go directly to the graph tab
	params.Add("g0.tab", "0")

	ret.RawQuery = params.Encode()
	ret.Path = "graph"

	return ret
}

func requestRateQuery(counterName, labelKey, labelValue string) string {
	return fmt.Sprintf("sum by (function, module) (rate(%s{%s=\"%s\"}[5m]))", counterName, labelKey, labelValue)
}

func errorRatioQuery(counterName, labelKey, labelValue string) string {
	return fmt.Sprintf("sum by (function, module) (rate(%s{%s=\"%s\",result=\"error\"}[5m]))", counterName, labelKey, labelValue)
}

func latencyQuery(bucketName, labelKey, labelValue string) string {
	latency := fmt.Sprintf("sum by (le, function, module) (rate(%s{%s=\"%s\"}[5m]))", bucketName, labelKey, labelValue)
	return fmt.Sprintf("histogram_quantile(0.99, %s) or histogram_quantile(0.95, %s)", latency, latency)
}

func concurrentCallsQuery(gaugeName, labelKey, labelValue string) string {
	return fmt.Sprintf("sum by (function, module) %s{%s=\"%s\"}", gaugeName, labelKey, labelValue)
}

func (p Prometheus) GenerateAutometricsComment(funcName string) []string {
	requestRateUrl := p.makePrometheusUrl(
		requestRateQuery(callCounterMetricName, "function", funcName), fmt.Sprintf("Rate of calls to the `%s` function per second, averaged over 5 minute windows", funcName))
	calleeRequestRateUrl := p.makePrometheusUrl(
		requestRateQuery(callCounterMetricName, "caller", funcName), fmt.Sprintf("Rate of function calls emanating from `%s` function per second, averaged over 5 minute windows", funcName))
	errorRatioUrl := p.makePrometheusUrl(
		errorRatioQuery(callCounterMetricName, "function", funcName), fmt.Sprintf("Percentage of calls to the `%s` function that return errors, averaged over 5 minute windows", funcName))
	calleeErrorRatioUrl := p.makePrometheusUrl(
		errorRatioQuery(callCounterMetricName, "caller", funcName), fmt.Sprintf("Percentage of function emanating from `%s` function that return errors, averaged over 5 minute windows", funcName))
	latencyUrl := p.makePrometheusUrl(
		latencyQuery(callDurationMetricName, "function", funcName), fmt.Sprintf("95th and 99th percentile latencies (in seconds) for the `%s` function", funcName))
	concurrentCallsUrl := p.makePrometheusUrl(
		concurrentCallsQuery(concurrentCallsMetricName, "function", funcName), fmt.Sprintf("Concurrent calls to the `%s` function", funcName))

	// Not using raw `` strings because it's impossible to escape ` within those
	return []string{
		"// ## Prometheus",
		"//",
		fmt.Sprintf("// View the live metrics for the `%s` function:", funcName),
		fmt.Sprintf("//   - [Request Rate](%s)", requestRateUrl.String()),
		fmt.Sprintf("//   - [Error Ratio](%s)", errorRatioUrl.String()),
		fmt.Sprintf("//   - [Latency (95th and 99th percentiles)](%s)", latencyUrl.String()),
		fmt.Sprintf("//   - [Concurrent Calls](%s)", concurrentCallsUrl.String()),
		"//",
		fmt.Sprintf("// Or, dig into the metrics of *functions called by* `%s`", funcName),
		fmt.Sprintf("//   - [Request Rate](%s)", calleeRequestRateUrl.String()),
		fmt.Sprintf("//   - [Error Ratio](%s)", calleeErrorRatioUrl.String()),
		"//",
	}
}
