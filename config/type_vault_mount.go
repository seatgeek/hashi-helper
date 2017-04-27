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
	Roles           []*MountRole
}

// MountInput ...
func (m *Mount) MountInput() *api.MountInput {
	return &api.MountInput{
		Type:        m.Type,
		Description: m.Description,
	}
}

// Mounts struct
//
// environment
type Mounts []*Mount

// Add ...
func (m *Mounts) Add(mount *Mount) {
	*m = append(*m, mount)
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
