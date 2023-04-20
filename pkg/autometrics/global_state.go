package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

var version string
var commit string
var buildTime string

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

// GetBuildTime returns the build timestamp of the codebase being instrumented.
func GetBuildTime() string {
	return buildTime
}

// SetBuildTime sets the build timestamp of the codebase being instrumented.
func SetBuildTime(newBuildTime string) {
	buildTime = newBuildTime
}
