package config

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
)

func (c *Config) processVaultPolicies(list *ast.ObjectList, environment *Environment, application *Application) error {
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
		printer.Fprint(buf, x.Children())
		policy.Raw = buf.String()

		// Replace environment and maybe app name placeholders
		policy.Raw = strings.Replace(policy.Raw, "__ENV__", environment.Name, -1)
		if application != nil {
			policy.Raw = strings.Replace(policy.Raw, "__APP__", application.Name, -1)
		}

		if err := hcl.DecodeObject(policy, policyAST); err != nil {
			return fmt.Errorf("Failed to parse policy: %s", err)
		}

		if o := list.Filter("path"); len(o.Items) > 0 {
			if err := parsePaths(policy, o); err != nil {
				return fmt.Errorf("Failed to parse policy: %s", err)
			}
		}

		if c.VaultPolicies.Add(policy) == false {
			if application != nil {
				c.logger.Warnf("      Ignored duplicate policy '%s' -> '%s' -> '%s' in line %s", environment.Name, application.Name, policy.Name, policyAST.Keys[0].Token.Pos)
			} else {
				c.logger.Warnf("      Ignored duplicate policy '%s' -> '%s' line %s", environment.Name, policy.Name, policyAST.Keys[0].Token.Type)
			}
		}
	}

	return nil
}
