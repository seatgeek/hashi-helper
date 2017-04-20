package config

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
)

func processPolicies(list *ast.ObjectList) (Policies, error) {
	policies := Policies{}

	if len(list.Items) > 0 {
		for _, policyAST := range list.Items {
			policyName := policyAST.Keys[0].Token.Value().(string)

			parsedPolicy, err := parsePolicy(policyAST.Val.(*ast.ObjectType).List)
			if err != nil {
				return nil, err
			}

			policies[policyName] = *parsedPolicy
		}
	}

	return policies, nil
}

// Parse is used to parse the specified ACL rules into an
// intermediary set of policies, before being compiled into
// the ACL
func parsePolicy(list *ast.ObjectList) (*Policy, error) {
	// Check for invalid top-level keys
	valid := []string{"name", "path"}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, fmt.Errorf("Failed to parse policy: %s", err)
	}

	// Create the initial policy and store the raw text of the rules
	var p Policy

	// Convert the HCL AST back to text so we can send it to the Vault API
	buf := new(bytes.Buffer)
	printer := printer.Config{}
	printer.Fprint(buf, list)
	p.Raw = buf.String()

	if err := hcl.DecodeObject(&p, list); err != nil {
		return nil, fmt.Errorf("Failed to parse policy: %s", err)
	}

	if o := list.Filter("path"); len(o.Items) > 0 {
		if err := parsePaths(&p, o); err != nil {
			return nil, fmt.Errorf("Failed to parse policy: %s", err)
		}
	}

	return &p, nil
}
