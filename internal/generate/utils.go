package generate // import "github.com/autometrics-dev/autometrics-go/internal/generate"

import (
	"strings"
)

// Backport of strings.CutPrefix for pre-1.20
func cutPrefix(s, prefix string) (after string, found bool) {
	if !strings.HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}

func filter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
