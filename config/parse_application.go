package config

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
)

func (c *Config) processApplications(applicationsAST *ast.ObjectList, environment *Environment) error {
	if len(applicationsAST.Items) > 1 {
		return fmt.Errorf("only one application stanza is allowed per file")
	}

	c.logger = c.logger.WithField("stanza", "application")
	c.logger.Debugf("Found %d application{}", len(applicationsAST.Items))
	for _, appAST := range applicationsAST.Items {
		if len(appAST.Keys) != 1 {
			return fmt.Errorf("Missing application name in line %+v", appAST.Keys[0].Pos())
		}

		appName := appAST.Keys[0].Token.Value().(string)

		if TargetApplication != "" && appName != TargetApplication {
			c.logger.Debugf("Skipping application %s (!= %s)", appName, TargetApplication)
			continue
		}

		// Check for valid keys inside an application stanza
		x := appAST.Val.(*ast.ObjectType).List
		valid := []string{"secret", "secrets", "policy", "kv"}
		if err := checkHCLKeys(x, valid); err != nil {
			return err
		}

		application := c.Applications.GetOrSet(&Application{Environment: environment, Name: appName})

		environment.Applications.Add(application)

		c.logger.Debug("Scanning for vault secret{}")
		if err := c.processVaultSecret(x.Filter("secret"), environment, application); err != nil {
			return err
		}

		c.logger.Debug("Scanning for vault secrets{}")
		if err := c.processVaultSecrets(x.Filter("secrets"), environment, application); err != nil {
			return err
		}

		c.logger.Debug("Scanning for policy")
		if err := c.processVaultPolicies(x.Filter("policy"), environment, application); err != nil {
			return err
		}

		c.logger.Debug("Scanning for consul KV")
		if err := c.processConsulKV(x.Filter("kv"), environment, application); err != nil {
			return err
		}

		c.Applications.Add(application)
	}

	return nil
}
