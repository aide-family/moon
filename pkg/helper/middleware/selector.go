package middleware

import (
	"context"
	"regexp"
	"strings"

	middle "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
)

type (
	// Selector is middleware selector
	Selector struct {
		ms     []middle.Middleware
		match  selector.MatchFunc
		client bool

		prefix []string
		regex  []string
		path   []string
	}

	transporter func(ctx context.Context) (transport.Transporter, bool)
)

// Server selector middleware
func Server(ms ...middle.Middleware) *Selector {
	return &Selector{ms: ms}
}

// Client selector middleware
func Client(ms ...middle.Middleware) *Selector {
	return &Selector{client: true, ms: ms}
}

var (
	// serverTransporter is get server transport.Transporter from ctx
	serverTransporter transporter = func(ctx context.Context) (transport.Transporter, bool) {
		return transport.FromServerContext(ctx)
	}

	// clientTransporter is get client transport.Transporter from ctx
	clientTransporter transporter = func(ctx context.Context) (transport.Transporter, bool) {
		return transport.FromClientContext(ctx)
	}
)

// Prefix is with Builder's prefix
func (b *Selector) Prefix(prefix ...string) *Selector {
	b.prefix = prefix
	return b
}

// Regex is with Builder's regex
func (b *Selector) Regex(regex ...string) *Selector {
	b.regex = regex
	return b
}

// Path is with Builder's path
func (b *Selector) Path(path ...string) *Selector {
	b.path = path
	return b
}

// Match is with Builder's match
func (b *Selector) Match(fn selector.MatchFunc) *Selector {
	b.match = fn
	return b
}

// Build is Builder's Build, for example: Server().Path(m1,m2).Build()
func (b *Selector) Build() middle.Middleware {
	var tf func(ctx context.Context) (transport.Transporter, bool)
	if b.client {
		tf = clientTransporter
	} else {
		tf = serverTransporter
	}
	return selectorFun(tf, b.matches, b.ms...)
}

type contextOperationKey struct{}

// GetOperation is get operation from context
func GetOperation(ctx context.Context) string {
	if v, ok := ctx.Value(contextOperationKey{}).(string); ok {
		return v
	}
	return ""
}

// matches is match operation compliance Builder
func (b *Selector) matches(ctx context.Context, transporter transporter) bool {
	info, ok := transporter(ctx)
	if !ok {
		return false
	}

	operation := info.Operation()
	ctx = context.WithValue(ctx, contextOperationKey{}, operation)
	for _, prefix := range b.prefix {
		if prefixMatch(prefix, operation) {
			return true
		}
	}
	for _, regex := range b.regex {
		if regexMatch(regex, operation) {
			return true
		}
	}
	for _, path := range b.path {
		if pathMatch(path, operation) {
			return true
		}
	}

	if b.match != nil {
		return b.match(ctx, operation)
	}

	return false
}

func pathMatch(path string, operation string) bool {
	return path == operation
}

func prefixMatch(prefix string, operation string) bool {
	return strings.HasPrefix(operation, prefix)
}

func regexMatch(regex string, operation string) bool {
	r, err := regexp.Compile(regex)
	if err != nil {
		return false
	}
	return r.FindString(operation) == operation
}

// selector middleware
func selectorFun(transporter transporter, match func(context.Context, transporter) bool, ms ...middle.Middleware) middle.Middleware {
	return func(handler middle.Handler) middle.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if !match(ctx, transporter) {
				return handler(ctx, req)
			}
			return middle.Chain(ms...)(handler)(ctx, req)
		}
	}
}
