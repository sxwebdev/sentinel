// Package rbacconnect provides a flexible, declarative Role-Based Access Control (RBAC)
// system for connect-go RPC services. It allows you to define fine-grained authorization
// rules based on procedures, services, packages, or default policies, with support for
// super-roles and dynamic policy updates.
//
// Key Features:
//   - Declarative policy builder with fluent API
//   - Multiple selector levels: Procedure > Service > Package > Default
//   - Super-roles that bypass all rules
//   - Allow/Deny rules with deny precedence
//   - Thread-safe atomic policy updates
//   - Connect-go interceptor integration
//
// Example usage:
//
//	policy := rbacconnect.NewPolicyBuilder().
//	    WithSuperRoles("root").
//	    WithDefaultAllow(false).
//	    When(rbacconnect.Proc("/api.UserService/GetUser")).
//	        Allow("admin", "user").
//	    When(rbacconnect.Svc("api.AdminService")).
//	        Allow("admin").
//	    When(rbacconnect.Pkg("api")).
//	        Deny("guest").
//	    Build()
//
//	provider := rbacconnect.NewProvider(policy)
//	interceptor := rbacconnect.NewInterceptor(provider, rbacconnect.Options{})
package rbacconnect

import "slices"

// Role represents a user role as a string.
// Roles are used to determine access permissions in the RBAC system.
type Role = string

// RoleSet is a set of roles implemented as a map for efficient lookup.
type RoleSet map[Role]struct{}

// NewRoleSet creates a new RoleSet from the provided roles.
func NewRoleSet(roles ...Role) RoleSet {
	rs := make(RoleSet, len(roles))
	for _, r := range roles {
		rs[r] = struct{}{}
	}
	return rs
}

// Has checks if the role exists in the set.
func (rs RoleSet) Has(r Role) bool {
	_, ok := rs[r]
	return ok
}

// Selector defines the scope of a rule application.
// Specificity order: Procedure > Service > Package > Default.
type Selector struct {
	// Procedure is the full procedure name like "/pkg.Service/Method".
	Procedure string
	// Service is the service name like "pkg.Service".
	Service string
	// Package is the package name like "pkg".
	Package string
	// Default indicates a default rule (when nothing else matches).
	Default bool
}

// Proc creates a selector for a specific procedure.
// Example: Proc("/api.UserService/GetUser")
func Proc(procedure string) Selector { return Selector{Procedure: procedure} }

// Svc creates a selector for all procedures in a service.
// Example: Svc("api.UserService")
func Svc(service string) Selector { return Selector{Service: service} }

// Pkg creates a selector for all procedures in a package.
// Example: Pkg("api")
func Pkg(pkg string) Selector { return Selector{Package: pkg} }

// Any creates a default selector that matches everything.
func Any() Selector { return Selector{Default: true} }

// Rule represents allow/deny permissions for roles.
// If a role is in Deny, it takes precedence over Allow.
type Rule struct {
	Allow RoleSet
	Deny  RoleSet
}

// Policy is a set of RBAC rules. Can be updated atomically.
type Policy struct {
	// Maps are stored separately for fast lookup by level.
	proc map[string]Rule
	svc  map[string]Rule
	pkg  map[string]Rule
	def  *Rule // default rule, optional

	// super contains super-roles that are always allowed regardless of rules
	super RoleSet
	// defaultAllow controls behavior when no rules match.
	defaultAllow bool
}

// Decision represents the result of an authorization check.
type Decision struct {
	Allowed  bool
	Selector string // which level matched: "procedure"|"service"|"package"|"default"|"none"
	Matched  string // which name matched
	Reason   string // description for logging
}

// Evaluate is a pure function that makes an authorization decision.
// It checks the given procedure, service, and package against the policy rules
// for the provided roles, returning a Decision.
func (p *Policy) Evaluate(procFull string, svc string, pkg string, roles []Role) Decision {
	// Check super-roles first
	for _, r := range roles {
		if p.super != nil && p.super.Has(r) {
			return Decision{Allowed: true, Selector: "super", Matched: r, Reason: "super-role"}
		}
	}

	// Check rules in order of specificity
	if d := evalRule(p.proc[procFull], roles); d != nil {
		return Decision{Allowed: *d, Selector: "procedure", Matched: procFull}
	}
	if d := evalRule(p.svc[svc], roles); d != nil {
		return Decision{Allowed: *d, Selector: "service", Matched: svc}
	}
	if d := evalRule(p.pkg[pkg], roles); d != nil {
		return Decision{Allowed: *d, Selector: "package", Matched: pkg}
	}
	if p.def != nil {
		if d := evalRule(*p.def, roles); d != nil {
			return Decision{Allowed: *d, Selector: "default", Matched: "*"}
		}
	}
	// fallback
	return Decision{Allowed: p.defaultAllow, Selector: "none", Matched: "", Reason: "fallback"}
}

// evalRule evaluates a single rule against the provided roles.
// Returns nil if no rule exists, true if allowed, false if denied.
func evalRule(r Rule, roles []Role) *bool {
	if len(r.Allow) == 0 && len(r.Deny) == 0 {
		return nil // "no rule"
	}

	// Explicit deny takes precedence
	if slices.ContainsFunc(roles, r.Deny.Has) {
		f := false
		return &f
	}

	// Allow if there's an intersection
	if slices.ContainsFunc(roles, r.Allow.Has) {
		t := true
		return &t
	}

	f := false
	return &f
}
