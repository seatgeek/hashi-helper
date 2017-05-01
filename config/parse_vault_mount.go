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

		valid := []string{"config", "role", "type", "path"}
		if err := checkHCLKeys(x, valid); err != nil {
			return err
		}

		if len(mountAST.Keys) != 1 {
			return fmt.Errorf("Missing mount name in line %+v", mountAST.Keys[0].Pos())
		}

		mountName := mountAST.Keys[0].Token.Value().(string)

		typeAST := x.Filter("type")
		if len(typeAST.Items) != 1 {
			return fmt.Errorf("missing mount type in %s -> %s", environment.Name, mountName)
		}

		mountType := typeAST.Items[0].Val.(*ast.LiteralType).Token.Value().(string)

		mount := &Mount{
			Name:        mountName,
			Type:        mountType,
			Environment: environment,
		}

		configAST := x.Filter("config")
		if len(configAST.Items) > 0 {
			config, err := c.parseMountConfig(configAST)
			if err != nil {
				return err
			}

			mount.Config = config
		}

		roleAST := x.Filter("role")
		if len(roleAST.Items) > 0 {
			roles, err := c.parseMountRole(roleAST)
			if err != nil {
				return err
			}

			mount.Roles = roles
		}

		c.VaultMounts.Add(mount)
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

func (c *Config) parseMountRole(list *ast.ObjectList) ([]*MountRole, error) {
	roles := make([]*MountRole, 0)

	for _, config := range list.Items {
		if len(config.Keys) < 1 {
			return nil, fmt.Errorf("Missing mount role name in line %+v", config.Keys[0].Pos())
		}

		var m map[string]interface{}
		if err := hcl.DecodeObject(&m, config.Val); err != nil {
			return nil, err
		}

		var role MountRole
		role.Name = config.Keys[0].Token.Value().(string)

		if err := mapstructure.WeakDecode(m, &role.Data); err != nil {
			return nil, err
		}

		roles = append(roles, &role)
	}

	return roles, nil
}
