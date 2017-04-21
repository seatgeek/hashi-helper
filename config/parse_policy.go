package config

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
)

func (c *Config) processPolicies(list *ast.ObjectList, environment *Environment, application *Application) error {
	if len(list.Items) < 1 {
		return nil
	}

	for _, policyAST := range list.Items {
		x := policyAST.Val.(*ast.ObjectType).List

		// Check for invalid top-level keys
		valid := []string{"name", "path"}
		if err := checkHCLKeys(x, valid); err != nil {
			return fmt.Errorf("Failed to parse policy: %s", err)
		}

		// Create the initial policy and store the raw text of the rules
		policy := &Policy{
			Name:        policyAST.Keys[0].Token.Value().(string),
			Environment: environment,
			Application: application,
		}

		// Convert the HCL AST back to text so we can send it to the Vault API
		buf := new(bytes.Buffer)
		printer := printer.Config{}
		printer.Fprint(buf, list)
		policy.Raw = buf.String()

		if err := hcl.DecodeObject(policy, policyAST); err != nil {
			return fmt.Errorf("Failed to parse policy: %s", err)
		}

		if o := list.Filter("path"); len(o.Items) > 0 {
			if err := parsePaths(policy, o); err != nil {
				return fmt.Errorf("Failed to parse policy: %s", err)
			}
		}

		c.Policies.Add(policy)
	}

	return nil
}
