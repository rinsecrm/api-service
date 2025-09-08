package canaryctx

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	CanaryHeader     = "X-Canary"
	CanaryHeaderGRPC = "x-canary"
)

type contextKey string

const canaryKey contextKey = "canary"

// FromContext extracts the canary PR number from context
func FromContext(ctx context.Context) (string, bool) {
	canary, ok := ctx.Value(canaryKey).(string)
	return canary, ok
}

// WithCanary adds canary PR number to context
func WithCanary(ctx context.Context, canary string) context.Context {
	return context.WithValue(ctx, canaryKey, canary)
}

// IsValidCanary checks if the canary value is a valid PR number (digits only)
func IsValidCanary(canary string) bool {
	if canary == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^\d+$`, canary)
	return matched
}

// HTTPMiddleware extracts X-Canary header from HTTP requests
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		canary := r.Header.Get(CanaryHeader)
		if canary != "" {
			canary = strings.TrimSpace(canary)
			if IsValidCanary(canary) {
				ctx := WithCanary(r.Context(), canary)
				r = r.WithContext(ctx)

				// Optionally echo the canary header back for debugging
				w.Header().Set("X-Canary-Echo", canary)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// UnaryClientInterceptor adds X-Canary to outgoing gRPC metadata
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if canary, ok := FromContext(ctx); ok {
			md := metadata.Pairs(CanaryHeaderGRPC, canary)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
