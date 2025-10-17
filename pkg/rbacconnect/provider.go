package rbacconnect

import "sync/atomic"

// Provider is a thread-safe atomic policy provider for "hot" updates.
// It allows updating the policy without restarting the service.
type Provider struct {
	cur atomic.Pointer[Policy]
}

// NewProvider creates a new Provider with the given initial policy.
func NewProvider(p *Policy) *Provider {
	pr := &Provider{}
	pr.cur.Store(p)
	return pr
}

// Get returns the current policy.
func (p *Provider) Get() *Policy {
	return p.cur.Load()
}

// Update atomically replaces the current policy with a new one.
func (p *Provider) Update(newP *Policy) {
	p.cur.Store(newP)
}
