package config

// Mount struct ...
type Mount struct {
	Name   string
	Type   string
	Config []*MountConfig
	Roles  []*MountRole
}

// Mounts struct
//
// environment
type Mounts map[string]Mount

// MountConfig ...
type MountConfig struct {
	Name string
	Data map[string]string
}

// MountRole ...
type MountRole struct {
	Name string
	Data map[string]string
}
