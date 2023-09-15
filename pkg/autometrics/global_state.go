package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

// These variables are describing the state of the application being autometricized,
// _not_ the build information of the binary

const (
	// AutometricsServiceNameEnv is the name of the environment variable to declare to fetch the name of
	// the service to use as a label. This environment variable has precedence over [OTelServiceNameEnv]
	// _and_ over hardcoding the variable directly in [BuildInfo] struct in the Init call.
	AutometricsServiceNameEnv = "AUTOMETRICS_SERVICE_NAME"
	// OTelServiceNameEnv is the name of the environment variable to declare to fetch the name of
	// the service to use as a label. This environment variable has precedence over variables hardcoded
	// in the [BuildInfo] struct in the Init call.
	OTelServiceNameEnv = "OTEL_SERVICE_NAME"
)

var (
	version     string
	commit      string
	branch      string
	service     string
	pushJobName string
	pushJobURL  string
)

// GetVersion returns the version of the codebase being instrumented.
func GetVersion() string {
	return version
}

// SetVersion sets the version of the codebase being instrumented.
func SetVersion(newVersion string) {
	version = newVersion
}

// GetCommit returns the commit of the codebase being instrumented.
func GetCommit() string {
	return commit
}

// SetCommit sets the commit of the codebase being instrumented.
func SetCommit(newCommit string) {
	commit = newCommit
}

// GetBranch returns the branch of the build of the codebase being instrumented.
func GetBranch() string {
	return branch
}

// SetBranch sets the branch of the build of the codebase being instrumented.
func SetBranch(newBranch string) {
	branch = newBranch
}

// GetService returns the service of the build of the codebase being instrumented.
func GetService() string {
	return service
}

// SetService sets the service name of the build of the codebase being instrumented.
func SetService(newService string) {
	service = newService
}

// GetPushJobName returns the job name to use when the codebase being instrumented is pushing metrics to an OTEL Collector.
func GetPushJobName() string {
	return pushJobName
}

// SetPushJobName sets the job name to use when the codebase being instrumented is pushing metrics to an OTEL Collector.
func SetPushJobName(newPushJobName string) {
	pushJobName = newPushJobName
}

// GetPushJobURL returns the job url to use when the codebase being instrumented is pushing metrics to an OTEL Collector.
func GetPushJobURL() string {
	return pushJobURL
}

// SetPushJobURL sets the job url to use when the codebase being instrumented is pushing metrics to an OTEL Collector.
func SetPushJobURL(newPushJobURL string) {
	pushJobURL = newPushJobURL
}
