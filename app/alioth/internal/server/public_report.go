package server

import (
	"context"
	"sync"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"

	v1 "github.com/aide-family/moon/api/alioth/v1"
)

const (
	publicReportRate  = 30.0
	publicReportBurst = 30.0
	publicReportMax   = 10000
)

type reportBucket struct {
	tokens float64
	last   time.Time
}

// publicReportLimiter 是进程内的公开报告访问限流。多副本时按副本分别限流，
// 只用于挡住按 token 的枚举与滥用，不承担精确配额。
type publicReportLimiter struct {
	mu      sync.Mutex
	buckets map[string]*reportBucket
}

func newPublicReportLimiter() *publicReportLimiter {
	return &publicReportLimiter{buckets: make(map[string]*reportBucket, 256)}
}

func (limiter *publicReportLimiter) allow(key string, now time.Time) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if len(limiter.buckets) > publicReportMax {
		for bucketKey := range limiter.buckets {
			delete(limiter.buckets, bucketKey)
		}
	}
	bucket, ok := limiter.buckets[key]
	if !ok {
		limiter.buckets[key] = &reportBucket{tokens: publicReportBurst - 1, last: now}
		return true
	}
	elapsed := now.Sub(bucket.last).Seconds()
	bucket.tokens += elapsed * publicReportRate
	if bucket.tokens > publicReportBurst {
		bucket.tokens = publicReportBurst
	}
	bucket.last = now
	if bucket.tokens < 1 {
		return false
	}
	bucket.tokens--
	return true
}

// publicReportGuard 给公开报告响应加 no-store / noindex，并按 token 限流。
func publicReportGuard() middleware.Middleware {
	limiter := newPublicReportLimiter()
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok || tr.Operation() != v1.OperationInspectionRunServiceGetSharedInspectionReport {
				return handler(ctx, req)
			}
			tr.ReplyHeader().Set("Cache-Control", "no-store")
			tr.ReplyHeader().Set("X-Robots-Tag", "noindex, nofollow")
			key := ""
			if request, ok := req.(*v1.GetSharedInspectionReportRequest); ok {
				key = request.GetToken()
			}
			if key == "" || !limiter.allow(key, time.Now()) {
				return nil, kerrors.New(429, "RATE_LIMITED", "too many requests")
			}
			return handler(ctx, req)
		}
	}
}
