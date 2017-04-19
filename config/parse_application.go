package config

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/hcl/hcl/ast"
)

func processApplications(applicationsAST *ast.ObjectList, environment Environment) (Applications, error) {
	applications := Applications{}

	if len(applicationsAST.Items) > 0 {
		for _, appAST := range applicationsAST.Items {
			if len(appAST.Keys) != 1 {
				return nil, fmt.Errorf("Missing application name in line %+v", appAST.Keys[0].Pos())
			}

			appName := appAST.Keys[0].Token.Value().(string)

			if TargetApplication != "" && appName != TargetApplication {
				log.Debugf("Skipping application %s (!= %s)", appName, TargetApplication)
				continue
			}

			app, err := parseApplication(appAST.Val.(*ast.ObjectType).List, environment.Name, appName)
			if err != nil {
				return nil, err
			}

			if _, ok := applications[appName]; !ok {
				applications[appName] = *app
			} else {
				applications[appName].merge(*app)
			}
		}
	}

	return applications, nil
}

// parseEnvironmentStanza
// parse out `environment -> application {)` stanza
func parseApplication(list *ast.ObjectList, envName, appName string) (*Application, error) {
	valid := []string{"secret", "policy"}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, err
	}

	application := Application{}

	secretsAST := list.Filter("secret")
	if len(secretsAST.Items) > 0 {
		secrets, err := parseSecretStanza(secretsAST, envName, appName)
		if err != nil {
			return nil, err
		}

		application.Secrets = secrets
	}

	policiesAST := list.Filter("policy")
	if len(policiesAST.Items) > 0 {
		policies, err := processPolicies(policiesAST)
		if err != nil {
			return nil, err
		}

		application.Policies = policies
	}

	return &application, nil
}
