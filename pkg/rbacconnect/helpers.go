package rbacconnect

// AllowProcs adds allow rules for the specified procedures and role.
func AllowProcs(b *PolicyBuilder, role Role, procs ...string) {
	for _, p := range procs {
		b.When(Proc(p)).Allow(role)
	}
}

// DenyProcs adds deny rules for the specified procedures and role.
func DenyProcs(b *PolicyBuilder, role Role, procs ...string) {
	for _, p := range procs {
		b.When(Proc(p)).Deny(role)
	}
}

// AllowServices adds allow rules for all procedures of the specified services and role.
func AllowServices(b *PolicyBuilder, role Role, services ...string) {
	for _, s := range services {
		b.When(Svc(s)).Allow(role)
	}
}

// DenyServices adds deny rules for all procedures of the specified services and role.
func DenyServices(b *PolicyBuilder, role Role, services ...string) {
	for _, s := range services {
		b.When(Svc(s)).Deny(role)
	}
}

// AllowPackages adds allow rules for all procedures of the specified packages and role.
func AllowPackages(b *PolicyBuilder, role Role, packages ...string) {
	for _, p := range packages {
		b.When(Pkg(p)).Allow(role)
	}
}

// DenyPackages adds deny rules for all procedures of the specified packages and role.
func DenyPackages(b *PolicyBuilder, role Role, packages ...string) {
	for _, p := range packages {
		b.When(Pkg(p)).Deny(role)
	}
}
