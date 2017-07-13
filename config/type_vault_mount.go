package config

import "github.com/hashicorp/vault/api"

// Mount struct ...
type Mount struct {
	Environment     *Environment
	Name            string
	Type            string
	Description     string
	DefaultLeaseTTL string
	MaxLeaseTTL     string
	Config          []*MountConfig
	Roles           MountRoles
}

// MountInput ...
func (m *Mount) MountInput() *api.MountInput {
	return &api.MountInput{
		Type:        m.Type,
		Description: m.Description,
	}
}

// MountRoles ...
type MountRoles []*MountRole

// Add ...
func (r *MountRoles) Add(role *MountRole) {
	*r = append(*r, role)
}

// VaultMounts struct
//
// environment
type VaultMounts []*Mount

// Add ...
func (m *VaultMounts) Add(mount *Mount) {
	*m = append(*m, mount)
}

// Find ...
func (m *VaultMounts) Find(name string) *Mount {
	for _, mount := range *m {
		if mount.Name == name {
			return mount
		}
	}

	return nil
}

// MountConfig ...
type MountConfig struct {
	Name string
	Data map[string]interface{}
}

// MountRole ...
type MountRole struct {
	Name string
	Data map[string]interface{}
}
