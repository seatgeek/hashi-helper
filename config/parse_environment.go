package config

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
)

// Parse root environment stanza
func (c *Config) processEnvironments(list *ast.ObjectList) error {
	valid := []string{"environment"}
	if err := checkHCLKeys(list, valid); err != nil {
		return err
	}

	environmentsAST := list.Filter("environment")
	if len(environmentsAST.Items) > 1 {
		return fmt.Errorf("only one environment stanza is allowed per file")
	}

	for _, envAST := range environmentsAST.Items {
		// ensure that we have a named environment
		//
		// aka 'environment "name" {}'
		// aka 'environment "name" "name2" {}'
		if len(envAST.Keys) == 0 {
			return fmt.Errorf("Missing environment name in: %+v", list.Pos())
		}

		for _, envKey := range envAST.Keys {
			// extract the name of the environment stanza
			envName := envKey.Token.Value().(string)
			c.logger.Debugf("  Found environment %s", envName)

			// check if we are limiting to a specific environment, and skip the current environment
			// if it does not match the required environment name
			if shouldSkipEnvironment(envName, TargetEnvironment) {
				c.logger.Debugf("  Skipping environment %s (%s != %s)", envName, envName, TargetEnvironment)
				continue
			}

			// Rewrite "*" to the actual target environment
			if envName == "*" {
				envName = TargetEnvironment
			}

			// check for valid keys inside an environment stanza
			x := envAST.Val.(*ast.ObjectType).List
			valid := []string{"application", "auth", "policy", "mount", "secret", "secrets", "service", "kv"}
			if err := checkHCLKeys(x, valid); err != nil {
				return err
			}

			env := c.Environments.GetOrSet(&Environment{Name: envName})

			c.logger.Debug("Scanning for applications")
			if err := c.processApplications(x.Filter("application"), env); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault auth{}")
			if err := c.processVaultAuths(x.Filter("auth"), env); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault secret{}")
			if err := c.processVaultSecret(x.Filter("secret"), env, nil); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault secrets{}")
			if err := c.processVaultSecrets(x.Filter("secrets"), env, nil); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault policy{}")
			if err := c.processVaultPolicies(x.Filter("policy"), env, nil); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault mount{}")
			if err := c.processVaultMounts(x.Filter("mount"), env); err != nil {
				return err
			}

			c.logger.Debug("Scanning for consul service{}")
			if err := c.processConsulServices(x.Filter("service"), env); err != nil {
				return err
			}

			c.logger.Debug("Scanning for consul kv{}")
			if err := c.processConsulKV(x.Filter("kv"), env, nil); err != nil {
				return err
			}

			c.Environments.Add(env)
		}

		// spew.Dump(c)
	}

	return nil
}

func shouldSkipEnvironment(parsedEnv, targetEnv string) bool {
	// * env mean it applies to any filtered environment
	if parsedEnv == "*" {
		return false
	}

	// no target env means skip everything
	if targetEnv == "" {
		return true
	}

	return targetEnv != parsedEnv
}
