package req

import (
	"math"
	"math/rand"
	"time"
)

func defaultGetRetryInterval(resp *Response, attempt int) time.Duration {
	return 100 * time.Millisecond
}

// RetryConditionFunc is a retry condition, which determines
// whether the request should retry.
type RetryConditionFunc func(resp *Response, err error) bool

// RetryHookFunc is a retry hook which will be executed before a retry.
type RetryHookFunc func(resp *Response, err error)

// GetRetryIntervalFunc is a function that determines how long should
// sleep between retry attempts.
type GetRetryIntervalFunc func(resp *Response, attempt int) time.Duration

func backoffInterval(min, max time.Duration) GetRetryIntervalFunc {
	base := float64(min)
	capLevel := float64(max)
	return func(resp *Response, attempt int) time.Duration {
		temp := math.Min(capLevel, base*math.Exp2(float64(attempt)))
		halfTemp := int64(temp / 2)
		sleep := halfTemp + rand.Int63n(halfTemp)
		return time.Duration(sleep)
	}
}

func newDefaultRetryOption() *retryOption {
	return &retryOption{
		GetRetryInterval: defaultGetRetryInterval,
	}
}

type retryOption struct {
	MaxRetries       int
	GetRetryInterval GetRetryIntervalFunc
	RetryConditions  []RetryConditionFunc
	RetryHooks       []RetryHookFunc
}

func (ro *retryOption) Clone() *retryOption {
	if ro == nil {
		return nil
	}
	o := &retryOption{
		MaxRetries:       ro.MaxRetries,
		GetRetryInterval: ro.GetRetryInterval,
	}
	for _, c := range ro.RetryConditions {
		o.RetryConditions = append(o.RetryConditions, c)
	}
	for _, h := range ro.RetryHooks {
		o.RetryHooks = append(o.RetryHooks, h)
	}
	return o
}
