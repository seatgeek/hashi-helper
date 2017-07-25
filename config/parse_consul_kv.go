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

		if len(kvAST.Keys) != 1 {
			return fmt.Errorf("Missing kv path in line %+v", kvAST.Keys[0].Pos())
		}

		key := kvAST.Keys[0].Token.Value().(string)

		value, err := getKeyString("value", x)
		if err != nil {
			return err
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
