package config

// Mount struct ...
type Mount struct {
	Environment *Environment
	Name        string
	Type        string
	Config      []*MountConfig
	Roles       []*MountRole
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
	Data map[string]string
}

// MountRole ...
type MountRole struct {
	Name string
	Data map[string]string
}
