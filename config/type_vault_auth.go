package config

import "github.com/hashicorp/vault/api"

// Auth struct ...
type Auth struct {
	Environment     *Environment
	Name            string
	Type            string
	Description     string
	DefaultLeaseTTL string
	MaxLeaseTTL     string
	Config          []*AuthConfig
	Roles           []*AuthRole
}

// AuthInput ...
func (m *Mount) AuthInput() *api.MountInput {
	return &api.MountInput{
		Type:        m.Type,
		Description: m.Description,
	}
}

// VaultAuths struct
//
// environment
type VaultAuths []*Auth

// Add ...
func (m *VaultAuths) Add(auth *Auth) {
	*m = append(*m, auth)
}

// AuthConfig ...
type AuthConfig struct {
	Name string
	Data map[string]interface{}
}

// AuthRole ...
type AuthRole struct {
	Name string
	Data map[string]interface{}
}
