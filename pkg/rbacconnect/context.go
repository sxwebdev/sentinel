package rbacconnect

import "context"

// RoleExtractor defines how to extract roles from a context.
// Users can provide their own implementation.
type RoleExtractor interface {
	ExtractRoles(ctx context.Context) ([]Role, error)
}

// RoleExtractorFunc is a function adapter for RoleExtractor interface.
type RoleExtractorFunc func(ctx context.Context) ([]Role, error)

// ExtractRoles implements the RoleExtractor interface.
func (f RoleExtractorFunc) ExtractRoles(ctx context.Context) ([]Role, error) {
	return f(ctx)
}

// Standard context-based extractor (optional).
type ctxKey struct{}

var rolesKey ctxKey

// WithRoles adds roles to the context.
func WithRoles(ctx context.Context, roles ...Role) context.Context {
	return context.WithValue(ctx, rolesKey, roles)
}

// RolesFromContext retrieves roles from the context.
func RolesFromContext(ctx context.Context) ([]Role, bool) {
	v := ctx.Value(rolesKey)
	if v == nil {
		return nil, false
	}
	if rr, ok := v.([]Role); ok {
		return rr, true
	}
	return nil, false
}
