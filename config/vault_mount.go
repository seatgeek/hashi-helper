package config

import (
	"fmt"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

// Mount struct ...
type Mount struct {
	Environment     *Environment
	Name            string
	Type            string
	Description     string
	DefaultLeaseTTL string
	MaxLeaseTTL     string
	ForceNoCache    bool
	Config          []*MountConfig
	Roles           MountRoles
}

// MountInput ...
func (m *Mount) MountInput() *api.MountInput {
	return &api.MountInput{
		Type:        m.Type,
		Description: m.Description,
		Config: api.MountConfigInput{
			DefaultLeaseTTL: m.DefaultLeaseTTL,
			MaxLeaseTTL:     m.MaxLeaseTTL,
			ForceNoCache:    m.ForceNoCache,
		},
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

func (c *Config) parseVaultMountStanza(list *ast.ObjectList, environment *Environment) error {
	if len(list.Items) == 0 {
		return nil
	}

	c.logger = c.logger.WithField("stanza", "mount")
	c.logger.Debugf("Found %d mount{}", len(list.Items))
	for _, mountAST := range list.Items {
		x := mountAST.Val.(*ast.ObjectType).List

		valid := []string{"config", "role", "type", "path", "max_lease_ttl", "default_lease_ttl", "force_no_cache"}
		if err := c.checkHCLKeys(x, valid); err != nil {
			return err
		}

		if len(mountAST.Keys) != 1 {
			return fmt.Errorf("Missing mount name in line %+v", mountAST.Keys[0].Pos())
		}

		mountName := mountAST.Keys[0].Token.Value().(string)

		mount := c.VaultMounts.Find(mountName)
		existing := true
		if mount == nil {
			existing = false
			typeAST := x.Filter("type")
			if len(typeAST.Items) != 1 {
				return fmt.Errorf("missing mount type in %s -> %s", environment.Name, mountName)
			}
			mountType := typeAST.Items[0].Val.(*ast.LiteralType).Token.Value().(string)

			mountMaxLeaseTTL := ""
			maxTTLAST := x.Filter("max_lease_ttl")
			if len(maxTTLAST.Items) == 1 {
				v := maxTTLAST.Items[0].Val.(*ast.LiteralType).Token.Value()
				switch t := v.(type) {
				default:
					return fmt.Errorf("unexpected type %T for %s -> %s -> max_lease_ttl", environment.Name, mountName, t)
				case string:
					mountMaxLeaseTTL = v.(string)
				}
			} else if len(maxTTLAST.Items) > 1 {
				return fmt.Errorf("You can only specify max_lease_ttl once per mount in %s -> %s", environment.Name, mountName)
			}

			mountDefaultLeaseTTL := ""
			defaultTTLAST := x.Filter("default_lease_ttl")
			if len(defaultTTLAST.Items) == 1 {
				v := defaultTTLAST.Items[0].Val.(*ast.LiteralType).Token.Value()
				switch t := v.(type) {
				default:
					return fmt.Errorf("unexpected type %T for %s -> %s -> default_lease_ttl", environment.Name, mountName, t)
				case string:
					mountDefaultLeaseTTL = v.(string)
				}
			} else if len(defaultTTLAST.Items) > 1 {
				return fmt.Errorf("You can only specify default_lease_ttl once per mount in %s -> %s", environment.Name, mountName)
			}

			mountForceNoCache := false
			forceNoCacheAST := x.Filter("force_no_cache")
			if len(forceNoCacheAST.Items) == 1 {
				v := forceNoCacheAST.Items[0].Val.(*ast.LiteralType).Token.Value()
				switch t := v.(type) {
				default:
					return fmt.Errorf("unexpected type %T for %s -> %s -> force_no_cache", environment.Name, mountName, t)
				case bool:
					mountForceNoCache = v.(bool)
				}
			} else if len(forceNoCacheAST.Items) > 1 {
				return fmt.Errorf("You can only specify force_no_cache once per mount in %s -> %s", environment.Name, mountName)
			}

			mount = &Mount{
				Name:            mountName,
				Type:            mountType,
				Environment:     environment,
				MaxLeaseTTL:     mountMaxLeaseTTL,
				DefaultLeaseTTL: mountDefaultLeaseTTL,
				ForceNoCache:    mountForceNoCache,
			}
		}

		configAST := x.Filter("config")
		if len(configAST.Items) > 0 && existing {
			return fmt.Errorf("You are modifying an existing mount (%s), can't change config", mountName)
		}

		if len(configAST.Items) > 0 {
			config, err := c.parseMountConfig(configAST)
			if err != nil {
				return err
			}

			mount.Config = config
		}

		roleAST := x.Filter("role")
		if len(roleAST.Items) > 0 {
			err := c.parseMountRole(roleAST, mount)
			if err != nil {
				return err
			}
		}

		if !existing {
			c.VaultMounts.Add(mount)
		}
	}

	return nil
}

func (c *Config) parseMountConfig(list *ast.ObjectList) ([]*MountConfig, error) {
	configs := make([]*MountConfig, 0)

	for _, mountConfigAST := range list.Items {
		if len(mountConfigAST.Keys) < 1 {
			return nil, fmt.Errorf("Missing mount role name in mount config stanza")
		}

		var m map[string]interface{}
		if err := hcl.DecodeObject(&m, mountConfigAST.Val); err != nil {
			return nil, err
		}

		var config MountConfig
		config.Name = mountConfigAST.Keys[0].Token.Value().(string)

		if err := mapstructure.WeakDecode(m, &config.Data); err != nil {
			return nil, err
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

func (c *Config) parseMountRole(list *ast.ObjectList, mount *Mount) error {
	for _, config := range list.Items {
		if len(config.Keys) < 1 {
			return fmt.Errorf("Missing mount role name in mount config stanza")
		}

		var m map[string]interface{}
		if err := hcl.DecodeObject(&m, config.Val); err != nil {
			return err
		}

		var role MountRole
		role.Name = config.Keys[0].Token.Value().(string)

		if err := mapstructure.WeakDecode(m, &role.Data); err != nil {
			return err
		}

		mount.Roles.Add(&role)
	}

	return nil
}
