package config

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/hcl/hcl/ast"
)

// ConsulKV ...
type ConsulKV struct {
	Application *Application
	Environment *Environment
	Key         string
	Value       []byte
}

// ToConsulKV ...
func (c *ConsulKV) ToConsulKV() *api.KVPair {
	return &api.KVPair{
		Key:   c.toPath(),
		Value: c.Value,
	}
}

func (c *ConsulKV) toPath() string {
	if c.Application != nil {
		return fmt.Sprintf("%v/%v", c.Application.Name, c.Key)
	}
	return c.Key
}

// ConsulKV struct
//
type ConsulKVs []*ConsulKV

// add ...
func (cs *ConsulKVs) add(kv *ConsulKV) {
	*cs = append(*cs, kv)
}

func (c *Config) parseConsulKVStanza(list *ast.ObjectList, env *Environment, app *Application) error {
	if len(list.Items) == 0 {
		return nil
	}

	c.logger.Debugf("Found %d kv{}", len(list.Items))
	for _, kvAST := range list.Items {
		x := kvAST.Val.(*ast.ObjectType).List

		valid := []string{"value"}
		if err := c.checkHCLKeys(x, valid); err != nil {
			return err
		}

		if len(kvAST.Keys) == 0 {
			return fmt.Errorf("Missing kv path in line %+v", kvAST.Keys[0].Pos())
		}

		key := kvAST.Keys[0].Token.Value().(string)

		var value string

		if len(kvAST.Keys) == 1 {
			var err error
			value, err = getKeyString("value", x)
			if err != nil {
				return err
			}
		} else if len(kvAST.Keys) == 2 {
			value = kvAST.Keys[1].Token.Value().(string)
		} else {
			return fmt.Errorf("Invalid number of parameter (%+v) on kv in line %+v", len(kvAST.Keys), kvAST.Keys[0].Pos())
		}

		kv := &ConsulKV{
			Application: app,
			Environment: env,
			Key:         key,
			Value:       []byte(value),
		}

		c.ConsulKVs.add(kv)
	}

	return nil
}
