package config

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
)

func (c *Config) processVaultMounts(list *ast.ObjectList, environment *Environment) error {
	if len(list.Items) == 0 {
		return nil
	}

	for _, mountAST := range list.Items {
		x := mountAST.Val.(*ast.ObjectType).List

		valid := []string{"config", "role", "type", "path", "maxleasettl"}
		if err := checkHCLKeys(x, valid); err != nil {
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

			mountmaxttl:=""
			max_ttl_AST := x.Filter("maxleasettl")

			if len(max_ttl_AST.Items) == 1 {
				mountmaxttl = max_ttl_AST.Items[0].Val.(*ast.LiteralType).Token.Value().(string)
			}

			mount = &Mount{
				Name:        mountName,
				Type:        mountType,
				Environment: environment,
				MaxLeaseTTL: mountmaxttl,
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
			return nil, fmt.Errorf("Missing mount role name in line %+v", mountConfigAST.Keys[0].Pos())
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
			return fmt.Errorf("Missing mount role name in line %+v", config.Keys[0].Pos())
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
