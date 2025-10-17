package rbacconnect

// PolicyBuilder is a convenient builder for declarative policies.
type PolicyBuilder struct {
	p Policy
}

// NewPolicyBuilder creates a new policy builder with empty rules.
func NewPolicyBuilder() *PolicyBuilder {
	return &PolicyBuilder{
		p: Policy{
			proc: make(map[string]Rule),
			svc:  make(map[string]Rule),
			pkg:  make(map[string]Rule),
			// def=nil => no default rule
		},
	}
}

// WithSuperRoles adds super-roles that bypass all authorization rules.
// Super-roles are always granted access regardless of any rules.
func (b *PolicyBuilder) WithSuperRoles(roles ...Role) *PolicyBuilder {
	if b.p.super == nil {
		b.p.super = NewRoleSet()
	}
	for _, r := range roles {
		b.p.super[r] = struct{}{}
	}
	return b
}

// WithDefaultAllow sets the default behavior when no rules match.
// If true, access is allowed by default. If false, access is denied by default.
func (b *PolicyBuilder) WithDefaultAllow(v bool) *PolicyBuilder {
	b.p.defaultAllow = v
	return b
}

// When starts a new rule clause for the given selector.
// Returns a clause that can be chained with Allow() and Deny() calls.
func (b *PolicyBuilder) When(s Selector) *clause {
	return &clause{b: b, s: s}
}

// Build finalizes the policy construction and returns the immutable Policy.
func (b *PolicyBuilder) Build() *Policy {
	// We don't make copies of maps (internal package), but could if needed.
	return &b.p
}

// clause is a helper type for chaining When(Selector).Allow(...).Deny(...) calls.
type clause struct {
	b *PolicyBuilder
	s Selector
}

// Allow grants access to the specified roles for the current selector.
// Returns the clause to allow further chaining.
func (c *clause) Allow(roles ...Role) *clause {
	r := c.curry()
	if r.Allow == nil {
		r.Allow = NewRoleSet()
	}
	for _, ro := range roles {
		r.Allow[ro] = struct{}{}
	}
	c.apply(r)
	return c
}

// Deny blocks access for the specified roles for the current selector.
// Deny takes precedence over Allow.
// Returns the clause to allow further chaining.
func (c *clause) Deny(roles ...Role) *clause {
	r := c.curry()
	if r.Deny == nil {
		r.Deny = NewRoleSet()
	}
	for _, ro := range roles {
		r.Deny[ro] = struct{}{}
	}
	c.apply(r)
	return c
}

// When allows continuing the chain with a new selector.
func (c *clause) When(s Selector) *clause {
	return c.b.When(s)
}

// Build finalizes the policy construction and returns the immutable Policy.
func (c *clause) Build() *Policy {
	return c.b.Build()
}

// curry retrieves the current rule for the selector.
func (c *clause) curry() Rule {
	switch {
	case c.s.Procedure != "":
		return c.b.p.proc[c.s.Procedure]
	case c.s.Service != "":
		return c.b.p.svc[c.s.Service]
	case c.s.Package != "":
		return c.b.p.pkg[c.s.Package]
	case c.s.Default:
		if c.b.p.def != nil {
			return *c.b.p.def
		}
		return Rule{}
	default:
		return Rule{}
	}
}

// apply stores the rule for the selector.
func (c *clause) apply(r Rule) {
	switch {
	case c.s.Procedure != "":
		c.b.p.proc[c.s.Procedure] = r
	case c.s.Service != "":
		c.b.p.svc[c.s.Service] = r
	case c.s.Package != "":
		c.b.p.pkg[c.s.Package] = r
	case c.s.Default:
		c.b.p.def = &r
	}
}
