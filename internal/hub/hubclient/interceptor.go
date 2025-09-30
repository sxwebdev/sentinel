package hubclient

import (
	"context"

	"connectrpc.com/connect"
)

// interceptor is a client-side Connect interceptor that injects the
// Authorization header into every outgoing request.
type interceptor struct {
	token string
}

// newInterceptor creates a new interceptor that will set
// Authorization: Sentinel <token> on all requests.
func newInterceptor(token string) connect.Interceptor { //nolint:ireturn
	return &interceptor{token: token}
}

// WrapUnary sets the Authorization header for unary RPCs.
func (i *interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc { //nolint:ireturn
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Set header only on client side; on handler side Header() returns request headers from client.
		if req.Spec().IsClient {
			req.Header().Set("Authorization", "Sentinel "+i.token)
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient sets the Authorization header for streaming RPCs (client-side).
func (i *interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc { //nolint:ireturn
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		// The request headers are sent with the first call to Send; set before that happens.
		conn.RequestHeader().Set("Authorization", "Sentinel "+i.token)
		return conn
	}
}

// WrapStreamingHandler is a no-op for client-side interceptor; just forward.
func (i *interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc { //nolint:ireturn
	return next
}
