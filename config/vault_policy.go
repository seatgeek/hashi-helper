package config

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
)

// Policy is used to represent the policy specified by
// an ACL configuration.
type Policy struct {
	Environment *Environment
	Application *Application
	Name        string              `hcl:"name"`
	Paths       []*PathCapabilities `hcl:"-"`
	Raw         string
}

// Equal ...
func (p *Policy) Equal(o *Policy) bool {
	// name must be same
	if p.Name != o.Name {
		return false
	}

	// environmet must same
	if p.Environment.Equal(o.Environment) == false {
		return false
	}

	// @todo check Application

	return true
}

// VaultPolicies ...
type VaultPolicies []*Policy

// Add ...
func (p *VaultPolicies) Add(policy *Policy) bool {
	if p.Exists(policy) == false {
		*p = append(*p, policy)
		return true
	}

	return false
}

// Exists ...
func (p *VaultPolicies) Exists(policy *Policy) bool {
	for _, existing := range *p {
		if policy.Equal(existing) {
			return true
		}
	}

	return false
}

func (c *Config) parseVaultPolicyStanza(list *ast.ObjectList, environment *Environment, application *Application) error {
	if len(list.Items) < 1 {
		return nil
	}

	c.logger = c.logger.WithField("stanza", "policy")
	c.logger.Debugf("Found %d policy{}", len(list.Items))
	for _, policyAST := range list.Items {
		x := policyAST.Val.(*ast.ObjectType).List

		// Check for invalid top-level keys
		valid := []string{"name", "path"}
		if err := c.checkHCLKeys(x, valid); err != nil {
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
			if err := c.parsePaths(policy, o); err != nil {
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
