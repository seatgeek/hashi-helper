package config

// Policy is used to represent the policy specified by
// an ACL configuration.
type Policy struct {
	Environment *Environment
	Application *Application
	Name        string              `hcl:"name"`
	Paths       []*PathCapabilities `hcl:"-"`
	Raw         string
}

// Equal ...
func (p *Policy) Equal(o *Policy) bool {
	// name must be same
	if p.Name != o.Name {
		return false
	}

	// environmet must same
	if p.Environment.Equal(o.Environment) == false {
		return false
	}

	// @todo check Application

	return true
}

// VaultPolicies ...
type VaultPolicies []*Policy

// Add ...
func (p *VaultPolicies) Add(policy *Policy) bool {
	if p.Exists(policy) == false {
		*p = append(*p, policy)
		return true
	}

	return false
}

// Exists ...
func (p *VaultPolicies) Exists(policy *Policy) bool {
	for _, existing := range *p {
		if policy.Equal(existing) {
			return true
		}
	}

	return false
}
