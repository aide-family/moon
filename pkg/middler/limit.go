package middler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/ips"
)

// IPRateLimitConfig configuration for IP-based rate limiting
type IPRateLimitConfig struct {
	sync.RWMutex
	// Default maximum requests per minute (used when operation has no specific configuration)
	DefaultMaxRequestsPerMinute int64
	// Cache client
	Cache cache.Cache
	// Whether rate limiting is enabled
	Enabled bool
	// Rate limiting configuration for different operations
	OperationLimits map[string]int64
}

// IPRateLimit middleware for IP-based rate limiting
func IPRateLimit(config *IPRateLimitConfig) middleware.Middleware {
	if !config.Enabled {
		return func(handler middleware.Handler) middleware.Handler {
			return handler
		}
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Get HTTP request
			httpReq, ok := http.RequestFromServerContext(ctx)
			if !ok {
				// If not an HTTP request, allow it to pass through
				return handler(ctx, req)
			}

			// Get client IP
			clientIP := ips.GetClientIP(httpReq)
			if clientIP == "" {
				// If unable to get IP, allow it to pass through
				return handler(ctx, req)
			}

			// Get operation
			operation := ""
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation = tr.Operation()
			}

			// Get rate limiting configuration based on operation
			maxRequests := config.getMaxRequestsForOperation(operation)
			if maxRequests < 0 {
				return handler(ctx, req)
			}

			// Generate cache key, including operation information
			key := fmt.Sprintf("rate_limit:ip:%s:op:%s:%s", clientIP, operation, time.Now().Format("2006-01-02-15-04"))

			// Check if limit is exceeded
			allowed, err := config.Cache.IncMax(ctx, key, maxRequests, time.Minute)
			if err != nil {
				// If cache error occurs, log it but allow request to pass through
				// Logging can be added here
				return handler(ctx, req)
			}

			if !allowed {
				// Limit exceeded, return error
				return nil, merr.ErrorBadRequest("rate limit exceeded for IP: %s, operation: %s", clientIP, operation)
			}

			// Within limit, continue processing request
			return handler(ctx, req)
		}
	}
}

// getMaxRequestsForOperation get rate limiting configuration for specific operation
func (c *IPRateLimitConfig) getMaxRequestsForOperation(operation string) int64 {
	c.RLock()
	defer c.RUnlock()
	if len(c.OperationLimits) == 0 {
		return c.DefaultMaxRequestsPerMinute
	}

	if limit, exists := c.OperationLimits[operation]; exists {
		return limit
	}

	return c.DefaultMaxRequestsPerMinute
}

// NewIPRateLimitConfig create IP rate limiting configuration
func NewIPRateLimitConfig(cache cache.Cache) *IPRateLimitConfig {
	return &IPRateLimitConfig{
		DefaultMaxRequestsPerMinute: -1, // Default -1 means no limit
		Cache:                       cache,
		Enabled:                     true,
		OperationLimits:             make(map[string]int64),
	}
}

// SetOperationLimit set rate limiting frequency for specific operation
func (c *IPRateLimitConfig) SetOperationLimit(operation string, maxRequestsPerMinute int64) *IPRateLimitConfig {
	c.Lock()
	defer c.Unlock()
	if c.OperationLimits == nil {
		c.OperationLimits = make(map[string]int64)
	}
	c.OperationLimits[operation] = maxRequestsPerMinute
	return c
}

// AppendOperationLimit append rate limiting frequency for specific operation
func (c *IPRateLimitConfig) AppendOperationLimit(limits map[string]int64) *IPRateLimitConfig {
	c.Lock()
	defer c.Unlock()
	if c.OperationLimits == nil {
		c.OperationLimits = make(map[string]int64)
	}
	for operation, maxRequestsPerMinute := range limits {
		c.OperationLimits[operation] = maxRequestsPerMinute
	}
	return c
}

// SetDefaultLimit set default rate limiting frequency
func (c *IPRateLimitConfig) SetDefaultLimit(maxRequestsPerMinute int64) *IPRateLimitConfig {
	c.Lock()
	defer c.Unlock()
	c.DefaultMaxRequestsPerMinute = maxRequestsPerMinute
	return c
}

// Enable enable rate limiting
func (c *IPRateLimitConfig) Enable() *IPRateLimitConfig {
	c.Lock()
	defer c.Unlock()
	c.Enabled = true
	return c
}

// Disable disable rate limiting
func (c *IPRateLimitConfig) Disable() *IPRateLimitConfig {
	c.Lock()
	defer c.Unlock()
	c.Enabled = false
	return c
}
