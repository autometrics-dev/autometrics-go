package doc

import (
	"fmt"
	"net/url"

	"github.com/autometrics-dev/autometrics-go/internal/ctx"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

const DefaultPrometheusInstanceUrl = "http://localhost:9090/"

type Prometheus struct {
	instanceUrl url.URL
}

// NewPrometheusDoc builds a documentation comment generator that creates Prometheus links.
//
// The document generator implements the AutometricsLinkCommentGenerator interface.
func NewPrometheusDoc(instanceUrl string) Prometheus {
	// No way to have a url.URL constant, so we reparse it here

	prometheusInstanceUrl, _ := url.Parse(instanceUrl)
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
	latency := fmt.Sprintf("sum by (le, function, module) (rate(%s_bucket{%s=\"%s\"}[5m]))", bucketName, labelKey, labelValue)

	return fmt.Sprintf("histogram_quantile(0.99, %s) or histogram_quantile(0.95, %s)", latency, latency)
}

func concurrentCallsQuery(gaugeName, labelKey, labelValue string) string {
	return fmt.Sprintf("sum by (function, module) %s{%s=\"%s\"}", gaugeName, labelKey, labelValue)
}

func (p Prometheus) GenerateAutometricsComment(ctx ctx.AutometricsGeneratorContext, funcName, moduleName string) []string {
	requestRateUrl := p.makePrometheusUrl(
		requestRateQuery(autometrics.FunctionCallsCountName, "function", funcName), fmt.Sprintf("Rate of calls to the `%s` function per second, averaged over 5 minute windows", funcName))
	calleeRequestRateUrl := p.makePrometheusUrl(
		requestRateQuery(autometrics.FunctionCallsCountName, "caller", fmt.Sprintf("%s.%s", moduleName, funcName)), fmt.Sprintf("Rate of function calls emanating from `%s` function per second, averaged over 5 minute windows", funcName))
	errorRatioUrl := p.makePrometheusUrl(
		errorRatioQuery(autometrics.FunctionCallsCountName, "function", funcName), fmt.Sprintf("Percentage of calls to the `%s` function that return errors, averaged over 5 minute windows", funcName))
	calleeErrorRatioUrl := p.makePrometheusUrl(
		errorRatioQuery(autometrics.FunctionCallsCountName, "caller", fmt.Sprintf("%s.%s", moduleName, funcName)), fmt.Sprintf("Percentage of function emanating from `%s` function that return errors, averaged over 5 minute windows", funcName))
	latencyUrl := p.makePrometheusUrl(
		latencyQuery(autometrics.FunctionCallsDurationName, "function", funcName), fmt.Sprintf("95th and 99th percentile latencies (in seconds) for the `%s` function", funcName))
	concurrentCallsUrl := p.makePrometheusUrl(
		concurrentCallsQuery(autometrics.FunctionCallsConcurrentName, "function", funcName), fmt.Sprintf("Concurrent calls to the `%s` function", funcName))

	// Not using raw `` strings because it's impossible to escape ` within those
	retval := []string{
		"// ## Prometheus",
		"//",
		fmt.Sprintf("// View the live metrics for the `%s` function:", funcName),
		"//   - [Request Rate]",
		"//   - [Error Ratio]",
		"//   - [Latency (95th and 99th percentiles)]",
	}
	if ctx.Ctx.TrackConcurrentCalls {
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

	if ctx.Ctx.TrackConcurrentCalls {
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
