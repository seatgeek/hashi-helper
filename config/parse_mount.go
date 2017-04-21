package config

import (
	"fmt"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
)

func processMounts(mountAST *ast.ObjectList, environment Environment) (Mounts, error) {
	mounts := Mounts{}

	if len(mountAST.Items) > 0 {
		for _, appAST := range mountAST.Items {
			if len(appAST.Keys) < 1 {
				return nil, fmt.Errorf("Missing mount name in line %+v", appAST.Keys[0].Pos())
			}

			if len(appAST.Keys) < 2 {
				return nil, fmt.Errorf("Missing mount type in line %+v", appAST.Keys[0].Pos())
			}

			mount, err := parseMount(appAST.Val.(*ast.ObjectType).List)
			if err != nil {
				return nil, err
			}

			mount.Name = appAST.Keys[0].Token.Value().(string)
			mount.Type = appAST.Keys[1].Token.Value().(string)

			if _, ok := mounts[mount.Name]; !ok {
				mounts[mount.Name] = *mount
			} else {
				// mounts[name].merge(*app)
			}
		}
	}

	return mounts, nil
}

// parseEnvironmentStanza
// parse out `environment -> application {)` stanza
func parseMount(list *ast.ObjectList) (*Mount, error) {
	valid := []string{"config", "role"}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, err
	}

	mount := Mount{}

	configAST := list.Filter("config")
	if len(configAST.Items) > 0 {
		config, err := parseMountConfig(configAST)
		if err != nil {
			return nil, err
		}

		mount.Config = config
	}

	roleAST := list.Filter("role")
	if len(roleAST.Items) > 0 {
		roles, err := parseMountRole(roleAST)
		if err != nil {
			return nil, err
		}

		mount.Roles = roles
	}

	return &mount, nil
}

func parseMountConfig(list *ast.ObjectList) ([]*MountConfig, error) {
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

func parseMountRole(list *ast.ObjectList) ([]*MountRole, error) {
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
