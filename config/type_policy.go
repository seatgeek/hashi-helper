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

// Policies ...
type Policies []*Policy

// Add ...
func (p *Policies) Add(policy *Policy) {
	*p = append(*p, policy)
}
