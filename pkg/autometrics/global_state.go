package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

// These variables are describing the state of the application being autometricized,
// _not_ the build information of the binary

var version string
var commit string
var branch string

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
