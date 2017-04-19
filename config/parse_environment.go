package config

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/hcl/hcl/ast"
)

// Parse root environment stanza
func processEnvironments(list *ast.ObjectList) (*Environments, error) {
	valid := []string{"environment"}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, err
	}

	environments := Environments{}

	for _, envAST := range list.Items {
		// ensure that we have a named environment
		//
		// aka 'environment "name" {}'
		if len(envAST.Keys) != 2 {
			return nil, fmt.Errorf("Missing environment name in line %+v", envAST.Keys[0].Pos())
		}

		// extract the name of the environment stanza
		envName := envAST.Keys[1].Token.Value().(string)

		// check if we are limiting to a specific environment, and skip the current environment
		// if it does not match the required environment name
		if TargetEnvironment != "" && envName != TargetEnvironment {
			log.Debugf("Skipping environment %s (%s != %s)", envName, envName, TargetEnvironment)
			continue
		}

		// delegate parsing of sub-stanza to a different parser
		res, err := parseEnvironment(envAST.Val.(*ast.ObjectType).List, envName)
		if err != nil {
			return nil, err
		}

		// either add the new environment, or merge it on top of an existing one
		if _, ok := environments[envName]; !ok {
			environments[envName] = *res
		} else {
			environments[envName].merge(*res)
		}
	}

	return &environments, nil
}

// parseEnvironment
// parse out `environment -> application {)` stanza
func parseEnvironment(list *ast.ObjectList, envName string) (*Environment, error) {
	valid := []string{"application", "policy"}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, err
	}

	env := Environment{}
	env.Name = envName

	applications, err := processApplications(list.Filter("application"), env)
	if err != nil {
		return nil, err
	}
	env.Applications = applications

	policies, err := processPolicies(list.Filter("policy"))
	if err != nil {
		return nil, err
	}
	env.Policies = policies

	return &env, nil
}
