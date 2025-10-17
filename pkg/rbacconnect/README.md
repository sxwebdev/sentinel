# rbacconnect

A flexible, declarative Role-Based Access Control (RBAC) system for [connect-go](https://github.com/connectrpc/connect-go) RPC services.

## Features

- 🎯 **Declarative Policy Builder** - Define authorization rules with a fluent, chainable API
- 🔍 **Multi-Level Selectors** - Control access at procedure, service, package, or default levels
- 👑 **Super-Roles** - Define roles that bypass all authorization rules
- ✅ **Allow/Deny Rules** - Fine-grained control with deny precedence over allow
- 🔄 **Atomic Policy Updates** - Thread-safe hot-reloading of policies without service restart
- 🔌 **Connect-go Integration** - Drop-in interceptor for connect-go services
- 🎨 **Customizable** - Pluggable role extraction, spec parsing, and error handling

## Installation

```bash
go get github.com/sxwebdev/sentinel/pkg/rbacconnect
```

## Quick Start

### 1. Define Your Policy

```go
package main

import (
    "github.com/sxwebdev/sentinel/pkg/rbacconnect"
)

func createPolicy() *rbacconnect.Policy {
    return rbacconnect.NewPolicyBuilder().
        // Super-roles bypass all rules
        WithSuperRoles("root").

        // Default deny (fail-closed security)
        WithDefaultAllow(false).

        // Allow specific procedure
        When(rbacconnect.Proc("/api.UserService/GetUser")).
            Allow("admin", "user").

        // Allow entire service
        When(rbacconnect.Svc("api.AdminService")).
            Allow("admin").

        // Allow entire package
        When(rbacconnect.Pkg("api.public")).
            Allow("user", "guest").

        // Explicit deny (takes precedence)
        When(rbacconnect.Proc("/api.UserService/DeleteUser")).
            Deny("user").

        // Default rule for everything else
        When(rbacconnect.Any()).
            Allow("admin").

        Build()
}
```

### 2. Create Provider and Interceptor

```go
package main

import (
    "github.com/sxwebdev/sentinel/pkg/rbacconnect"
    "connectrpc.com/connect"
)

func main() {
    // Create policy
    policy := createPolicy()

    // Create provider (for atomic updates)
    provider := rbacconnect.NewProvider(policy)

    // Create interceptor
    interceptor := rbacconnect.NewInterceptor(provider, rbacconnect.Options{
        // Optional: custom role extractor
        RoleExtractor: rbacconnect.RoleExtractorFunc(func(ctx context.Context) ([]rbacconnect.Role, error) {
            // Extract roles from your auth context
            // Example: from JWT claims, session, etc.
            user := getUserFromContext(ctx)
            return user.Roles, nil
        }),
    })

    // Add interceptor to your connect-go server
    mux := http.NewServeMux()
    mux.Handle(greetv1connect.NewGreetServiceHandler(
        &greetServer{},
        connect.WithInterceptors(interceptor),
    ))
}
```

### 3. Add Roles to Context

```go
// In your authentication middleware
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract user and their roles
        user := authenticateUser(r)

        // Add roles to context
        ctx := rbacconnect.WithRoles(r.Context(), user.Roles...)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Rule Evaluation Order

Rules are evaluated in order of specificity, from most specific to least:

1. **Super-Roles** - Always granted access
2. **Procedure** - Exact procedure match (e.g., `/api.UserService/GetUser`)
3. **Service** - Service-level match (e.g., `api.UserService`)
4. **Package** - Package-level match (e.g., `api`)
5. **Default** - Explicit default rule via `When(rbacconnect.Any())`
6. **Fallback** - `WithDefaultAllow(true/false)`

Within each level, **Deny takes precedence over Allow**.

## Examples

### Example 1: Simple Service Authorization

```go
policy := rbacconnect.NewPolicyBuilder().
    WithSuperRoles("admin").
    WithDefaultAllow(false).
    When(rbacconnect.Svc("api.UserService")).
        Allow("user", "moderator").
    Build()
```

### Example 2: Public and Private APIs

```go
policy := rbacconnect.NewPolicyBuilder().
    WithDefaultAllow(false).

    // Public APIs - anyone can access
    When(rbacconnect.Pkg("api.public")).
        Allow("guest", "user", "admin").

    // Private APIs - authenticated users only
    When(rbacconnect.Pkg("api.private")).
        Allow("user", "admin").

    // Admin APIs - admins only
    When(rbacconnect.Pkg("api.admin")).
        Allow("admin").

    Build()
```

### Example 3: Mixed Allow/Deny Rules

```go
policy := rbacconnect.NewPolicyBuilder().
    WithDefaultAllow(false).

    // Users can access most of UserService
    When(rbacconnect.Svc("api.UserService")).
        Allow("user").

    // But users cannot delete
    When(rbacconnect.Proc("/api.UserService/DeleteUser")).
        Deny("user").
        Allow("admin").

    Build()
```

### Example 4: Hot-Reloading Policies

```go
// Initial policy
policy := createPolicy()
provider := rbacconnect.NewProvider(policy)

// Later, update the policy without restarting
newPolicy := rbacconnect.NewPolicyBuilder().
    WithSuperRoles("root", "superadmin").
    WithDefaultAllow(false).
    When(rbacconnect.Any()).
        Allow("admin").
    Build()

// Atomic update - all new requests use the new policy
provider.Update(newPolicy)
```

### Example 5: Custom Role Extractor from JWT

```go
import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    Roles []string `json:"roles"`
    jwt.RegisteredClaims
}

interceptor := rbacconnect.NewInterceptor(provider, rbacconnect.Options{
    RoleExtractor: rbacconnect.RoleExtractorFunc(func(ctx context.Context) ([]rbacconnect.Role, error) {
        // Extract JWT from context
        token := getTokenFromContext(ctx)
        if token == "" {
            return nil, errors.New("no token")
        }

        // Parse and validate JWT
        claims := &Claims{}
        _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
            return []byte("your-secret-key"), nil
        })
        if err != nil {
            return nil, err
        }

        return claims.Roles, nil
    }),
})
```

## API Reference

### Types

- **`Role`** - String type representing a user role
- **`RoleSet`** - Set of roles for efficient lookup
- **`Selector`** - Defines the scope of a rule (Procedure/Service/Package/Default)
- **`Rule`** - Allow/Deny permissions for roles
- **`Policy`** - Complete set of RBAC rules
- **`Decision`** - Result of authorization check

### Builder Methods

- **`NewPolicyBuilder()`** - Create a new policy builder
- **`WithSuperRoles(roles ...Role)`** - Add super-roles
- **`WithDefaultAllow(bool)`** - Set default behavior
- **`When(Selector)`** - Start a rule clause
- **`Allow(roles ...Role)`** - Grant access to roles
- **`Deny(roles ...Role)`** - Block access for roles
- **`Build()`** - Finalize and return the policy

### Selectors

- **`Proc(procedure string)`** - Match specific procedure
- **`Svc(service string)`** - Match all procedures in a service
- **`Pkg(package string)`** - Match all procedures in a package
- **`Any()`** - Match everything (default rule)

### Provider

- **`NewProvider(policy *Policy)`** - Create atomic policy provider
- **`Get()`** - Get current policy
- **`Update(newPolicy *Policy)`** - Atomically update policy

### Interceptor

- **`NewInterceptor(provider *Provider, opts Options)`** - Create RBAC interceptor

### Context Helpers

- **`WithRoles(ctx, roles ...Role)`** - Add roles to context
- **`RolesFromContext(ctx)`** - Extract roles from context

## Testing

```go
func TestPolicy(t *testing.T) {
    policy := rbacconnect.NewPolicyBuilder().
        WithSuperRoles("root").
        WithDefaultAllow(false).
        When(rbacconnect.Proc("/api.UserService/GetUser")).
            Allow("user").
        Build()

    // Test super-role
    dec := policy.Evaluate("/api.UserService/GetUser", "api.UserService", "api", []string{"root"})
    assert.True(t, dec.Allowed)

    // Test allowed role
    dec = policy.Evaluate("/api.UserService/GetUser", "api.UserService", "api", []string{"user"})
    assert.True(t, dec.Allowed)

    // Test denied role
    dec = policy.Evaluate("/api.UserService/GetUser", "api.UserService", "api", []string{"guest"})
    assert.False(t, dec.Allowed)
}
```

## Best Practices

1. **Fail Closed** - Use `WithDefaultAllow(false)` for security
2. **Least Privilege** - Grant minimum necessary permissions
3. **Explicit Deny** - Use deny rules for critical restrictions
4. **Test Policies** - Write unit tests for your authorization logic
5. **Audit Decisions** - Log authorization decisions for security auditing
6. **Version Policies** - Track policy changes in version control

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This package is part of the Sentinel project.

## Related Projects

- [connect-go](https://github.com/connectrpc/connect-go) - Simple, reliable RPC
- [connect-grpchealth-go](https://github.com/connectrpc/grpchealth-go) - gRPC health checks for connect

## Support

For issues and questions, please open an issue on GitHub.
