package autometrics // import "github.com/autometrics-dev/autometrics-go/internal/autometrics"

import (
	"fmt"
	"net/url"

	prometheus "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

type AutometricsLinkCommentGenerator interface {
	GenerateAutometricsComment(ctx GeneratorContext, funcName, moduleName string) []string
	// Generated Links returns the list of text from links created by
	// GenerateAutometricsComment.
	//
	// As gofmt will move the links outside the fences for Autometrics
	// documentation, we add a helper to track the links to delete/recreate
	// in the links section when calling the generator multiple times.
	GeneratedLinks() []string
}

type Prometheus struct {
	instanceUrl url.URL
}

// NewPrometheusDoc builds a documentation comment generator that creates Prometheus links.
//
// The document generator implements the AutometricsLinkCommentGenerator interface.
func NewPrometheusDoc(instanceUrl url.URL) Prometheus {
	// No way to have a url.URL constant, so we reparse it here

	return Prometheus{instanceUrl: instanceUrl}
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

func addBuildInfoLabels() string {
	return fmt.Sprintf("* on (instance, job) group_left(%s, %s) last_over_time(%s[1s])",
		prometheus.VersionLabel,
		prometheus.CommitLabel,
		prometheus.BuildInfoName,
	)
}

func requestRateQuery(counterName, labelKey, labelValue string) string {
	return fmt.Sprintf("sum by (%s, %s, %s, %s, %s) (rate(%s{%s=\"%s\"}[5m]) %s)",
		prometheus.FunctionLabel,
		prometheus.ModuleLabel,
		prometheus.ServiceNameLabel,
		prometheus.VersionLabel,
		prometheus.CommitLabel,
		counterName,
		labelKey,
		labelValue,
		addBuildInfoLabels(),
	)
}

func errorRatioQuery(counterName, labelKey, labelValue string) string {
	return fmt.Sprintf("(sum by (%s, %s, %s, %s, %s) (rate(%s{%s=\"%s\",%s=\"error\"}[5m]) %s)) / (%s)",
		prometheus.FunctionLabel,
		prometheus.ModuleLabel,
		prometheus.ServiceNameLabel,
		prometheus.VersionLabel,
		prometheus.CommitLabel,
		counterName,
		labelKey,
		labelValue,
		prometheus.ResultLabel,
		addBuildInfoLabels(),
		requestRateQuery(counterName, labelKey, labelValue),
	)
}

func latencyQuery(bucketName, labelKey, labelValue string) string {
	latency := fmt.Sprintf("sum by (le, %s, %s, %s, %s, %s) (rate(%s_bucket{%s=\"%s\"}[5m]) %s)",
		prometheus.FunctionLabel,
		prometheus.ModuleLabel,
		prometheus.ServiceNameLabel,
		prometheus.VersionLabel,
		prometheus.CommitLabel,
		bucketName,
		labelKey,
		labelValue,
		addBuildInfoLabels(),
	)

	return fmt.Sprintf(
		"label_replace(histogram_quantile(0.99, %s), \"percentile_latency\", \"99\", \"\", \"\") or "+
			"label_replace(histogram_quantile(0.95, %s),\"percentile_latency\", \"95\", \"\", \"\")",
		latency,
		latency,
	)
}

func concurrentCallsQuery(gaugeName, labelKey, labelValue string) string {
	return fmt.Sprintf("sum by (%s, %s, %s, %s, %s) (%s{%s=\"%s\"} %s)",
		prometheus.FunctionLabel,
		prometheus.ModuleLabel,
		prometheus.ServiceNameLabel,
		prometheus.VersionLabel,
		prometheus.CommitLabel,
		gaugeName,
		labelKey,
		labelValue,
		addBuildInfoLabels(),
	)
}

func (p Prometheus) GenerateAutometricsComment(ctx GeneratorContext, funcName, moduleName string) []string {
	requestRateUrl := p.makePrometheusUrl(
		requestRateQuery(prometheus.FunctionCallsCountName, prometheus.FunctionLabel, funcName), fmt.Sprintf("Rate of calls to the `%s` function per second, averaged over 5 minute windows", funcName))
	calleeRequestRateUrl := p.makePrometheusUrl(
		requestRateQuery(prometheus.FunctionCallsCountName, prometheus.CallerFunctionLabel, funcName), fmt.Sprintf("Rate of function calls emanating from `%s` function per second, averaged over 5 minute windows", funcName))
	errorRatioUrl := p.makePrometheusUrl(
		errorRatioQuery(prometheus.FunctionCallsCountName, prometheus.FunctionLabel, funcName), fmt.Sprintf("Percentage of calls to the `%s` function that return errors, averaged over 5 minute windows", funcName))
	calleeErrorRatioUrl := p.makePrometheusUrl(
		errorRatioQuery(prometheus.FunctionCallsCountName, prometheus.CallerFunctionLabel, funcName), fmt.Sprintf("Percentage of function emanating from `%s` function that return errors, averaged over 5 minute windows", funcName))
	latencyUrl := p.makePrometheusUrl(
		latencyQuery(prometheus.FunctionCallsDurationName, prometheus.FunctionLabel, funcName), fmt.Sprintf("95th and 99th percentile latencies (in seconds) for the `%s` function", funcName))
	concurrentCallsUrl := p.makePrometheusUrl(
		concurrentCallsQuery(prometheus.FunctionCallsConcurrentName, prometheus.FunctionLabel, funcName), fmt.Sprintf("Concurrent calls to the `%s` function", funcName))

	// Not using raw `` strings because it's impossible to escape ` within those
	retval := []string{
		"// # Prometheus",
		"//",
		fmt.Sprintf("// View the live metrics for the `%s` function:", funcName),
		"//   - [Request Rate]",
		"//   - [Error Ratio]",
		"//   - [Latency (95th and 99th percentiles)]",
	}
	if ctx.RuntimeCtx.TrackConcurrentCalls {
		retval = append(retval,
			"//   - [Concurrent Calls]",
		)
	}
	retval = append(retval,
		"//",
		fmt.Sprintf("// Or, dig into the metrics of *functions called by* `%s`", funcName),
		"//   - [Request Rate Callee]",
		"//   - [Error Ratio Callee]",
	)

	retval = append(retval,
		"//",
		fmt.Sprintf("// [Request Rate]: %s", requestRateUrl.String()),
		fmt.Sprintf("// [Error Ratio]: %s", errorRatioUrl.String()),
		fmt.Sprintf("// [Latency (95th and 99th percentiles)]: %s", latencyUrl.String()),
	)

	if ctx.RuntimeCtx.TrackConcurrentCalls {
		retval = append(retval,
			fmt.Sprintf("// [Concurrent Calls]: %s", concurrentCallsUrl.String()),
		)
	}

	retval = append(retval,
		fmt.Sprintf("// [Request Rate Callee]: %s", calleeRequestRateUrl.String()),
		fmt.Sprintf("// [Error Ratio Callee]: %s", calleeErrorRatioUrl.String()),
		"//",
	)

	return retval
}

func (p Prometheus) GeneratedLinks() []string {
	return []string{"Request Rate", "Error Ratio", "Latency (95th and 99th percentiles)", "Concurrent Calls", "Request Rate Callee", "Error Ratio Callee"}
}
