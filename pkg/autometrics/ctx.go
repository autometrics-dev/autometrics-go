package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

import (
	"context"
	"log"
	"math/rand"
	"time"
)

type contextKey int

const (
	currentTraceIdKey contextKey = iota
	currentSpanIdKey
	currentParentSpanIdKey
	currentTrackConcurrentCallsKey
	currentTrackCallerNameKey
	currentAlertConfigurationKey
	currentStartTimeKey
	currentCallInfoKey
	currentBuildInfoKey
	currentValidHttpCodeRangesKey
)

var randSource *rand.Rand

// Open Telemetry-compatible trace ID
type TraceID [16]byte

// Open Telemetry-compatible span ID
type SpanID [8]byte

// NewContext is a constructor taking the parent context as argument.
//
// It accepts 'nil' as the parent context. In this case the constructor
// acts as if it received a new, fresh context.Background().
func NewContext(parentCtx context.Context) context.Context {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx := SetTrackConcurrentCalls(parentCtx, true)
	ctx = SetTrackCallerName(ctx, true)
	ctx = SetValidHttpCodeRanges(ctx, []InclusiveIntRange{{Min: 100, Max: 399}})
	return ctx
}

// SetTrackConcurrentCalls sets a flag in the context deciding whether to track how many concurrent calls the instrumented functions observe.
//
// TrackConcurrentCalls triggers the collection of the gauge for concurrent calls of the function.
// The flag defaults to true.
func SetTrackConcurrentCalls(ctx context.Context, track bool) context.Context {
	return context.WithValue(ctx, currentTrackConcurrentCallsKey, track)
}

// GetTrackConcurrentCalls returns whether autometrics should track how many concurrent calls the instrumented function observe.
//
// TrackConcurrentCalls triggers the collection of the gauge for concurrent calls of the function.
// It defaults to true.
func GetTrackConcurrentCalls(c context.Context) bool {
	if c == nil {
		return true
	}

	track, ok := c.Value(currentTrackConcurrentCallsKey).(bool)
	if !ok {
		return true
	}

	return track
}

// SetTrackCallerName sets a flag in the context deciding whether to track the names of the callers of instrumented functions.
//
// TrackCallerName adds a label with the caller name in all the collected metrics.
// The flag defaults to true.
func SetTrackCallerName(ctx context.Context, track bool) context.Context {
	return context.WithValue(ctx, currentTrackCallerNameKey, track)
}

// GetTrackCallerName returns default information if the context did not contain any build information.
//
// TrackCallerName adds a label with the caller name in all the collected metrics.
// It defaults to true.
func GetTrackCallerName(c context.Context) bool {
	if c == nil {
		return true
	}

	track, ok := c.Value(currentTrackCallerNameKey).(bool)
	if !ok {
		return true
	}

	return track
}

// SetAlertConfiguration sets the context's [AlertConfiguration]
//
// AlertConfiguration is an optional configuration to add alerting capabilities to the metrics.
func SetAlertConfiguration(ctx context.Context, slo AlertConfiguration) context.Context {
	return context.WithValue(ctx, currentAlertConfigurationKey, slo)
}

// GetAlertConfiguration returns default information if the context did not contain any alerting configuration.
//
// AlertConfiguration is an optional configuration to add alerting capabilities to the metrics.
func GetAlertConfiguration(c context.Context) AlertConfiguration {
	if c == nil {
		return AlertConfiguration{}
	}

	slo, ok := c.Value(currentAlertConfigurationKey).(AlertConfiguration)
	if !ok {
		return AlertConfiguration{}
	}

	return slo
}

// SetCallInfo sets the context's [CallInfo]
//
// CallInfo contains all the relevant data for caller information.
func SetCallInfo(ctx context.Context, build CallInfo) context.Context {
	return context.WithValue(ctx, currentCallInfoKey, build)
}

// GetCallInfo returns default information if the context did not contain any build information.
//
// CallInfo contains all the relevant data for caller information.
func GetCallInfo(c context.Context) CallInfo {
	if c == nil {
		return CallInfo{}
	}

	build, ok := c.Value(currentCallInfoKey).(CallInfo)
	if !ok {
		return CallInfo{}
	}

	return build
}

// SetStartTime sets the context's [StartTime]
//
// StartTime is the start time of a single function execution.
func SetStartTime(ctx context.Context, newStartTime time.Time) context.Context {
	return context.WithValue(ctx, currentStartTimeKey, newStartTime)
}

// GetStartTime returns default current time if the context did not contain any start time.
//
// StartTime is the start time of a single function execution.
func GetStartTime(c context.Context) time.Time {
	if c == nil {
		return time.Now()
	}

	readStartTime, ok := c.Value(currentStartTimeKey).(time.Time)
	if !ok {
		log.Printf("Warning: startTime is not a time.")
		return time.Now()
	}

	return readStartTime
}

// SetBuildInfo sets the context's [BuildInfo]
//
// BuildInfo contains all the relevant data for caller information.
func SetBuildInfo(ctx context.Context, build BuildInfo) context.Context {
	return context.WithValue(ctx, currentBuildInfoKey, build)
}

// GetBuildInfo returns default information if the context did not contain any build information.
//
// BuildInfo contains all the relevant data for caller information.
func GetBuildInfo(c context.Context) BuildInfo {
	if c == nil {
		return BuildInfo{}
	}

	build, ok := c.Value(currentBuildInfoKey).(BuildInfo)
	if !ok {
		return BuildInfo{}
	}

	return build
}

// SetTraceID sets the context's [TraceID]
func SetTraceID(ctx context.Context, tid TraceID) context.Context {
	return context.WithValue(ctx, currentTraceIdKey, tid)
}

// GetTraceID returns (_, false) if the context did not contain any trace id.
func GetTraceID(c context.Context) (TraceID, bool) {
	if c == nil {
		return TraceID{}, false
	}
	tid, ok := c.Value(currentTraceIdKey).(TraceID)
	return tid, ok
}

// SetSpanID sets the context's [SpanID]
func SetSpanID(ctx context.Context, sid SpanID) context.Context {
	return context.WithValue(ctx, currentSpanIdKey, sid)
}

// GetSpanID returns (_, false) if the context did not contain the current span id.
func GetSpanID(c context.Context) (SpanID, bool) {
	if c == nil {
		return SpanID{}, false
	}
	sid, ok := c.Value(currentSpanIdKey).(SpanID)
	return sid, ok
}

// SetParentSpanID sets the context's span's parent [SpanID]
func SetParentSpanID(ctx context.Context, sid SpanID) context.Context {
	return context.WithValue(ctx, currentParentSpanIdKey, sid)
}

// GetParentSpanID returns (_, false) if the context did not contain the parent's span id (including when we are in the root span).
func GetParentSpanID(c context.Context) (SpanID, bool) {
	if c == nil {
		return SpanID{}, false
	}
	sid, ok := c.Value(currentParentSpanIdKey).(SpanID)
	return sid, ok
}

// FillTracingInfo ensures the context has a traceID and a spanID.
// If they do not have this information, this method adds randomly
// generated IDs in the context to be used later for exemplars
//
// The random generator is a PRNG, seeded with the timestamp of the first time new IDs are needed.
func FillTracingInfo(ctx context.Context) context.Context {
	// We are using a PRNG because FillTracingInfo is expected to be called in PreInstrument.
	// Therefore it can have a noticeable impact on the performance of instrumented code.
	// Pseudo randomness should be enough for our use cases, true randomness might introduce too much latency.
	// randSource is initialized with a timestamp from the first time it is accessed in nanoseconds, which should
	// be enough precision to avoid accidental collisions (imagine multiple services starting "at the same time" in a deployment).
	if randSource == nil {
		randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	if parentSpanId, ok := GetSpanID(ctx); ok {
		ctx = SetParentSpanID(ctx, parentSpanId)
	}

	sid := SpanID{}
	_, _ = randSource.Read(sid[:])
	ctx = SetSpanID(ctx, sid)

	if _, ok := GetTraceID(ctx); !ok {
		tid := TraceID{}
		_, _ = randSource.Read(tid[:])
		ctx = SetTraceID(ctx, tid)
	}

	return ctx
}

// GenerateTraceId generates a new TraceID with a Pseudo-random number generator.
//
// The generator is seeded with the timestamp of the first time a new traceID is needed.
func GenerateTraceId() TraceID {
	if randSource == nil {
		randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	tid := TraceID{}
	randSource.Read(tid[:])

	return tid
}

// WithNewTraceId returns a copy of the passed context, with a newly generated traceID accessible for autometrics.
func WithNewTraceId(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, currentTraceIdKey, GenerateTraceId())
}

// FillBuildInfo adds the relevant build information to the current context.
func FillBuildInfo(ctx context.Context) context.Context {
	b := BuildInfo{
		Version: GetVersion(),
		Commit:  GetCommit(),
		Branch:  GetBranch(),
		Service: GetService(),
	}

	return SetBuildInfo(ctx, b)
}

type InclusiveIntRange struct {
	Min int
	Max int
}

func (r InclusiveIntRange) Contains(value int) bool {
	return value >= r.Min && value <= r.Max
}

// SetValidHttpCodeRanges sets the values of http codes that Autometrics should consider as "ok" results on calls.
//
// The value to set is an array of `(int, int)` pairs, where each pair contains the inclusive minimum and inclusive maximum of a valid range.
// For example, `[(100,399)]` is a good default, if we only want 4xx and 5xx status codes to be errors for Autometrics reporting. Another
// option that might be popular is `[(100, 499)]` to only see server-side errors as errors in Autometrics metrics/dashboards.
//
// The ability to specify multiple ranges allow for disjoint sets: `[(100, 399), (418,418)]` would make Autometrics report an error on
// any 4xx or 5xx status code, _except_ if that code is 418 (I'm a teapot).
//
// This setting is only useful when used in conjunction with the [github.com/autometrics-dev/autometrics-go/pkg/middleware/http/middleware.Autometrics] wrapper.
func SetValidHttpCodeRanges(ctx context.Context, ranges []InclusiveIntRange) context.Context {
	return context.WithValue(ctx, currentParentSpanIdKey, ranges)
}

// GetValidHttpCodeRanges returns the list of values that should be considered as "ok" by Autometrics when computing the success rate of a handler.
//
// Look at the documentation of [SetValidHttpCodeRanges] for more information about the semantics of the returned value.
func GetValidHttpCodeRanges(c context.Context) []InclusiveIntRange {
	if c == nil {
		return []InclusiveIntRange{{
			Min: 100,
			Max: 399,
		}}
	}

	ranges, ok := c.Value(currentParentSpanIdKey).([]InclusiveIntRange)
	if !ok {
		return []InclusiveIntRange{}
	}

	return ranges
}
