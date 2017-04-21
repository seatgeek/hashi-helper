package config

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/hcl/hcl/ast"
)

func (c *Config) processApplications(applicationsAST *ast.ObjectList, environment *Environment) error {
	if len(applicationsAST.Items) == 0 {
		return nil
	}

	for _, appAST := range applicationsAST.Items {
		if len(appAST.Keys) != 1 {
			return fmt.Errorf("Missing application name in line %+v", appAST.Keys[0].Pos())
		}

		appName := appAST.Keys[0].Token.Value().(string)

		if TargetApplication != "" && appName != TargetApplication {
			log.Debugf("Skipping application %s (!= %s)", appName, TargetApplication)
			continue
		}

		// Check for valid keys inside an application stanza
		x := appAST.Val.(*ast.ObjectType).List
		valid := []string{"secret", "policy"}
		if err := checkHCLKeys(x, valid); err != nil {
			return err
		}

		application := c.Applications.GetOrSet(&Application{Environment: environment, Name: appName})

		environment.Applications.Add(application)

		log.Debug("    Scanning for secrets")
		if err := c.processSecrets(x.Filter("secret"), application); err != nil {
			return err
		}

		log.Debug("    Scanning for policy")
		if err := c.processPolicies(x.Filter("policy"), environment, application); err != nil {
			return err
		}

		c.Applications.Add(application)
	}

	return nil
}
