// Package rbacconnect provides a flexible, declarative Role-Based Access Control (RBAC)
// system for connect-go RPC services.
//
// # Overview
//
// rbacconnect enables fine-grained authorization for connect-go services through a
// declarative policy builder. It supports multiple levels of selectors (procedure,
// service, package, default), super-roles, allow/deny rules, and atomic policy updates.
//
// # Quick Start
//
// Define a policy:
//
//	policy := rbacconnect.NewPolicyBuilder().
//	    WithSuperRoles("root").
//	    WithDefaultAllow(false).
//	    When(rbacconnect.Proc("/api.UserService/GetUser")).
//	        Allow("admin", "user").
//	    When(rbacconnect.Svc("api.AdminService")).
//	        Allow("admin").
//	    Build()
//
// Create an interceptor:
//
//	provider := rbacconnect.NewProvider(policy)
//	interceptor := rbacconnect.NewInterceptor(provider, rbacconnect.Options{})
//
// Add to your connect-go server:
//
//	mux.Handle(yourservice.NewHandler(
//	    &yourServer{},
//	    connect.WithInterceptors(interceptor),
//	))
//
// # Policy Structure
//
// Policies are built using a fluent API with the following components:
//
//   - Super-Roles: Roles that bypass all rules (e.g., "root", "superadmin")
//   - Selectors: Define the scope of rules (Procedure, Service, Package, or Default)
//   - Rules: Allow or Deny specific roles (Deny takes precedence)
//   - Default Behavior: What happens when no rules match
//
// # Rule Evaluation Order
//
// Rules are evaluated in order of specificity:
//
//  1. Super-Roles (always allowed)
//  2. Procedure-level rules (most specific)
//  3. Service-level rules
//  4. Package-level rules
//  5. Default rule
//  6. Fallback to WithDefaultAllow setting
//
// Within each level, Deny takes precedence over Allow.
//
// # Selectors
//
// Four types of selectors are available:
//
//   - Proc(procedure): Match exact procedure like "/api.UserService/GetUser"
//   - Svc(service): Match all procedures in a service like "api.UserService"
//   - Pkg(package): Match all procedures in a package like "api"
//   - Any(): Match everything (default rule)
//
// # Hot Reloading
//
// Policies can be updated atomically without service restart:
//
//	newPolicy := rbacconnect.NewPolicyBuilder().
//	    WithSuperRoles("root", "admin").
//	    Build()
//	provider.Update(newPolicy)
//
// # Custom Role Extraction
//
// Implement RoleExtractor to customize how roles are extracted from context:
//
//	interceptor := rbacconnect.NewInterceptor(provider, rbacconnect.Options{
//	    RoleExtractor: rbacconnect.RoleExtractorFunc(func(ctx context.Context) ([]rbacconnect.Role, error) {
//	        // Extract roles from JWT, session, etc.
//	        return extractRolesFromJWT(ctx)
//	    }),
//	})
//
// # Thread Safety
//
// All components are thread-safe. The Provider uses atomic operations for policy updates,
// ensuring no race conditions during hot-reloads.
//
// # Best Practices
//
//   - Use WithDefaultAllow(false) for fail-closed security
//   - Grant minimum necessary permissions (principle of least privilege)
//   - Use explicit Deny for critical restrictions
//   - Test policies thoroughly with unit tests
//   - Log authorization decisions for security auditing
//   - Version control your policies
//
// # File Organization
//
// The package is organized into logical files:
//
//   - rbacconnect.go: Core types (Role, Selector, Rule, Policy)
//   - builder.go: PolicyBuilder and fluent API
//   - interceptor.go: Connect-go integration
//   - context.go: Context helpers and role extraction
//   - provider.go: Atomic policy provider
//   - helpers.go: Utility functions for common patterns
//   - utils.go: Internal utilities
//
// For more examples and detailed documentation, see the README.md file.
package rbacconnect
