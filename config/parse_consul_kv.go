package config

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
)

func (c *Config) processConsulKV(list *ast.ObjectList, env *Environment, app *Application) error {
	if len(list.Items) == 0 {
		return nil
	}

	for _, kvAST := range list.Items {
		x := kvAST.Val.(*ast.ObjectType).List

		valid := []string{"value"}
		if err := checkHCLKeys(x, valid); err != nil {
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

		c.ConsulKVs.Add(kv)
	}

	return nil
}
