package rbacconnect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
)

// Example 1: Basic policy with super-roles
func ExamplePolicyBuilder_basic() {
	policy := NewPolicyBuilder().
		WithSuperRoles("root").
		WithDefaultAllow(false).
		When(Proc("/api.UserService/GetUser")).
		Allow("admin", "user").
		Build()

	// Evaluate for admin role
	decision := policy.Evaluate("/api.UserService/GetUser", "api.UserService", "api", []string{"admin"})
	println(decision.Allowed) // true
}

// Example 2: Service-level authorization
func ExamplePolicyBuilder_service() {
	policy := NewPolicyBuilder().
		WithDefaultAllow(false).
		When(Svc("api.UserService")).
		Allow("user").
		When(Svc("api.AdminService")).
		Allow("admin").
		Build()

	// User can access UserService
	decision := policy.Evaluate("/api.UserService/GetUser", "api.UserService", "api", []string{"user"})
	println(decision.Allowed) // true

	// User cannot access AdminService
	decision = policy.Evaluate("/api.AdminService/DeleteUser", "api.AdminService", "api", []string{"user"})
	println(decision.Allowed) // false
}

// Example 3: Package-level with deny rules
func ExamplePolicyBuilder_denyRules() {
	policy := NewPolicyBuilder().
		WithDefaultAllow(false).
		// Allow users to access the package
		When(Pkg("api")).
		Allow("user", "admin").
		// But deny users from specific dangerous operations
		When(Proc("/api.UserService/DeleteUser")).
		Deny("user").
		Build()

	// User can read
	decision := policy.Evaluate("/api.UserService/GetUser", "api.UserService", "api", []string{"user"})
	println(decision.Allowed) // true

	// User cannot delete (deny takes precedence)
	decision = policy.Evaluate("/api.UserService/DeleteUser", "api.UserService", "api", []string{"user"})
	println(decision.Allowed) // false
}

// Example 4: Hot-reloading policies
func ExampleProvider_update() {
	// Initial policy
	initialPolicy := NewPolicyBuilder().
		WithDefaultAllow(false).
		When(Any()).
		Allow("admin").
		Build()

	provider := NewProvider(initialPolicy)

	// Check with initial policy
	policy := provider.Get()
	decision := policy.Evaluate("/any/Procedure", "any.Service", "any", []string{"user"})
	println(decision.Allowed) // false

	// Update policy
	newPolicy := NewPolicyBuilder().
		WithDefaultAllow(false).
		When(Any()).
		Allow("admin", "user"). // now users are allowed
		Build()

	provider.Update(newPolicy)

	// Check with new policy
	policy = provider.Get()
	decision = policy.Evaluate("/any/Procedure", "any.Service", "any", []string{"user"})
	println(decision.Allowed) // true
}

// Example 5: Custom role extractor
func ExampleNewInterceptor_customRoleExtractor() {
	policy := NewPolicyBuilder().
		WithDefaultAllow(false).
		When(Any()).
		Allow("admin").
		Build()

	provider := NewProvider(policy)

	// Custom role extractor from HTTP headers
	extractorFromHeader := RoleExtractorFunc(func(ctx context.Context) ([]Role, error) {
		// In real scenario, extract from actual context
		header := ctx.Value("X-User-Roles")
		if header == nil {
			return nil, errors.New("no roles header")
		}
		// Parse roles from header
		rolesStr, ok := header.(string)
		if !ok {
			return nil, errors.New("invalid roles")
		}
		return []string{rolesStr}, nil
	})

	_ = NewInterceptor(provider, Options{
		RoleExtractor: extractorFromHeader,
	})
}

// Example 6: Using context helpers
func ExampleWithRoles() {
	ctx := context.Background()

	// Add roles to context
	ctx = WithRoles(ctx, "admin", "user")

	// Extract roles from context
	roles, ok := RolesFromContext(ctx)
	if ok {
		println("Roles:", roles) // ["admin", "user"]
	}
}

// Example 7: Testing authorization logic
func TestPolicyEvaluation(t *testing.T) {
	policy := NewPolicyBuilder().
		WithSuperRoles("root").
		WithDefaultAllow(false).
		When(Proc("/api.UserService/GetUser")).
		Allow("user").
		When(Proc("/api.UserService/DeleteUser")).
		Allow("admin").
		Build()

	tests := []struct {
		name      string
		procedure string
		service   string
		pkg       string
		roles     []string
		wantAllow bool
	}{
		{
			name:      "root bypasses all rules",
			procedure: "/api.UserService/DeleteUser",
			service:   "api.UserService",
			pkg:       "api",
			roles:     []string{"root"},
			wantAllow: true,
		},
		{
			name:      "user can get",
			procedure: "/api.UserService/GetUser",
			service:   "api.UserService",
			pkg:       "api",
			roles:     []string{"user"},
			wantAllow: true,
		},
		{
			name:      "user cannot delete",
			procedure: "/api.UserService/DeleteUser",
			service:   "api.UserService",
			pkg:       "api",
			roles:     []string{"user"},
			wantAllow: false,
		},
		{
			name:      "admin can delete",
			procedure: "/api.UserService/DeleteUser",
			service:   "api.UserService",
			pkg:       "api",
			roles:     []string{"admin"},
			wantAllow: true,
		},
		{
			name:      "no roles - deny",
			procedure: "/api.UserService/GetUser",
			service:   "api.UserService",
			pkg:       "api",
			roles:     []string{},
			wantAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := policy.Evaluate(tt.procedure, tt.service, tt.pkg, tt.roles)
			if decision.Allowed != tt.wantAllow {
				t.Errorf("Evaluate() allowed = %v, want %v (reason: %s)",
					decision.Allowed, tt.wantAllow, decision.Reason)
			}
		})
	}
}

// Example 8: Integration with connect-go service
func ExampleInterceptor_withConnectService() {
	// Create policy
	policy := NewPolicyBuilder().
		WithSuperRoles("root").
		WithDefaultAllow(false).
		When(Pkg("greet.v1")).
		Allow("user").
		Build()

	provider := NewProvider(policy)
	_ = NewInterceptor(provider, Options{})

	// Create HTTP handler with interceptor
	mux := http.NewServeMux()

	// In real code, use your actual service handler:
	// mux.Handle(greetv1connect.NewGreetServiceHandler(
	//     &greetServer{},
	//     connect.WithInterceptors(interceptor),
	// ))

	// Start server
	_ = httptest.NewServer(mux)
}

// Example 9: Multi-level authorization
func ExamplePolicyBuilder_multiLevel() {
	policy := NewPolicyBuilder().
		WithDefaultAllow(false).
		// Package level - all users can access public APIs
		When(Pkg("api.public")).
		Allow("guest", "user", "admin").
		// Service level - only registered users for private APIs
		When(Svc("api.private.UserService")).
		Allow("user", "admin").
		// Procedure level - only admins for dangerous operations
		When(Proc("/api.private.UserService/DeleteAll")).
		Allow("admin").
		// Default - admins get everything else
		When(Any()).
		Allow("admin").
		Build()

	// Guest can access public
	d := policy.Evaluate("/api.public.InfoService/GetInfo", "api.public.InfoService", "api.public", []string{"guest"})
	println("Guest public:", d.Allowed) // true

	// Guest cannot access private
	d = policy.Evaluate("/api.private.UserService/GetProfile", "api.private.UserService", "api.private", []string{"guest"})
	println("Guest private:", d.Allowed) // false

	// User can access private
	d = policy.Evaluate("/api.private.UserService/GetProfile", "api.private.UserService", "api.private", []string{"user"})
	println("User private:", d.Allowed) // true

	// User cannot delete all
	d = policy.Evaluate("/api.private.UserService/DeleteAll", "api.private.UserService", "api.private", []string{"user"})
	println("User delete:", d.Allowed) // false

	// Admin can delete all
	d = policy.Evaluate("/api.private.UserService/DeleteAll", "api.private.UserService", "api.private", []string{"admin"})
	println("Admin delete:", d.Allowed) // true
}

// Example 10: Custom error factory
func ExampleNewInterceptor_customErrorFactory() {
	policy := NewPolicyBuilder().
		WithDefaultAllow(false).
		Build()

	provider := NewProvider(policy)

	interceptor := NewInterceptor(provider, Options{
		ErrorFactory: func(code connect.Code, msg string) error {
			// Custom error with additional context
			return connect.NewError(code, errors.New(msg+" - contact support"))
		},
	})

	_ = interceptor
}
