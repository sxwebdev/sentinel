package rbacconnect

import (
	"context"

	"connectrpc.com/connect"
)

// Options configures the Interceptor behavior.
type Options struct {
	// RoleExtractor extracts roles from the request context.
	RoleExtractor RoleExtractor
	// SpecSplitter parses connect.Spec into (proc, svc, pkg).
	// Default implementation parses "/pkg.Service/Method" format.
	SpecSplitter func(spec connect.Spec) (proc string, svc string, pkg string)
	// ErrorFactory creates custom errors (PermissionDenied/Unauthenticated).
	ErrorFactory func(code connect.Code, msg string) error
}

// Interceptor is a connect-go interceptor that enforces RBAC policies.
type Interceptor struct {
	provider *Provider
	opts     Options
}

// NewInterceptor creates a new RBAC interceptor with the given provider and options.
// If options are not provided, sensible defaults are used.
func NewInterceptor(provider *Provider, opts Options) *Interceptor {
	if opts.RoleExtractor == nil {
		opts.RoleExtractor = RoleExtractorFunc(func(ctx context.Context) ([]Role, error) {
			if r, ok := RolesFromContext(ctx); ok {
				return r, nil
			}
			return nil, Err("roles missing")
		})
	}
	if opts.SpecSplitter == nil {
		opts.SpecSplitter = defaultSpecSplitter
	}
	if opts.ErrorFactory == nil {
		opts.ErrorFactory = defaultErrorFactory
	}
	return &Interceptor{provider: provider, opts: opts}
}

// WrapUnary wraps unary RPC calls with RBAC authorization.
func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		roles, err := i.opts.RoleExtractor.ExtractRoles(ctx)
		if err != nil || len(roles) == 0 {
			return nil, i.opts.ErrorFactory(connect.CodeUnauthenticated, "rbac: no roles")
		}
		proc, svc, pkg := i.opts.SpecSplitter(req.Spec())
		p := i.provider.Get()
		dec := p.Evaluate(proc, svc, pkg, roles)
		if !dec.Allowed {
			return nil, i.opts.ErrorFactory(connect.CodePermissionDenied, "rbac: access denied")
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient wraps client-side streaming calls.
// By default, client streams are not restricted.
func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next // typically we don't restrict client streams
}

// WrapStreamingHandler wraps server-side streaming calls with RBAC authorization.
func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		roles, err := i.opts.RoleExtractor.ExtractRoles(ctx)
		if err != nil || len(roles) == 0 {
			return i.opts.ErrorFactory(connect.CodeUnauthenticated, "rbac: no roles")
		}
		proc, svc, pkg := i.opts.SpecSplitter(conn.Spec())
		p := i.provider.Get()
		dec := p.Evaluate(proc, svc, pkg, roles)
		if !dec.Allowed {
			return i.opts.ErrorFactory(connect.CodePermissionDenied, "rbac: access denied")
		}
		return next(ctx, conn)
	}
}

// defaultSpecSplitter parses a connect.Spec into procedure, service, and package names.
// Expected format: "/pkg.Service/Method"
func defaultSpecSplitter(spec connect.Spec) (string, string, string) {
	// spec.Procedure: "/pkg.Service/Method"
	proc := spec.Procedure
	// Extract svc "pkg.Service" and pkg "pkg"
	var svc, pkg string
	// Strip leading slash
	s := proc
	if len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	// s = "pkg.Service/Method"
	// svc — everything before the slash
	for i := 0; i < len(s); i++ {
		if s[i] == '/' { // found
			svc = s[:i]
			break
		}
	}
	// pkg — everything before the first dot
	for i := 0; i < len(svc); i++ {
		if svc[i] == '.' {
			pkg = svc[:i]
			break
		}
	}
	return proc, svc, pkg
}

// defaultErrorFactory creates a connect error with the given code and message.
func defaultErrorFactory(code connect.Code, msg string) error {
	return connect.NewError(code, Err(msg))
}

// Err is a simple error type for string errors.
type Err string

// Error implements the error interface.
func (e Err) Error() string {
	return string(e)
}
