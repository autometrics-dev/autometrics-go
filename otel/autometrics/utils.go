package autometrics // import "github.com/autometrics-dev/autometrics-go/otel/autometrics"

import (
	"context"
	"encoding/hex"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

// Convenience re-export of [hex.DecodeString] to allow generating code without touching imports in instrumented file.
func DecodeString(s string) []byte {
	res, err := hex.DecodeString(s)
	if err != nil {
		return nil
	}
	return res
}

// Convenience re-export of [autometrics.WithNewTraceId] to avoid needing multiple imports in instrumented file.
func WithNewTraceId(ctx context.Context) context.Context {
	return autometrics.WithNewTraceId(ctx)
}
