package config

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
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
		if len(envAST.Keys) != 1 {
			return fmt.Errorf("Missing environment name in line %+v", envAST.Keys[0].Pos())
		}

		// extract the name of the environment stanza
		envName := envAST.Keys[0].Token.Value().(string)
		log.Debugf("  Found environment %s", envName)

		// check if we are limiting to a specific environment, and skip the current environment
		// if it does not match the required environment name
		if TargetEnvironment != "" && envName != TargetEnvironment {
			log.Debugf("  Skipping environment %s (%s != %s)", envName, envName, TargetEnvironment)
			continue
		}

		// check for valid keys inside an environment stanza
		x := envAST.Val.(*ast.ObjectType).List
		valid := []string{"application", "auth", "policy", "mount", "secret", "service", "kv"}
		if err := checkHCLKeys(x, valid); err != nil {
			return err
		}

		env := c.Environments.GetOrSet(&Environment{Name: envName})

		log.Debug("  Scanning for applications")
		if err := c.processApplications(x.Filter("application"), env); err != nil {
			return err
		}

		log.Debug("  Scanning for vault auth backends")
		if err := c.processVaultAuths(x.Filter("auth"), env); err != nil {
			return err
		}

		log.Debug("  Scanning for vault secrets")
		if err := c.processVaultSecrets(x.Filter("secret"), env, nil); err != nil {
			return err
		}

		log.Debug("  Scanning for vault policies")
		if err := c.processVaultPolicies(x.Filter("policy"), env, nil); err != nil {
			return err
		}

		log.Debug("  Scanning for vault mounts")
		if err := c.processVaultMounts(x.Filter("mount"), env); err != nil {
			return err
		}

		log.Debug("  Scanning for consul services")
		if err := c.processConsulServices(x.Filter("service"), env); err != nil {
			return err
		}

		log.Debug("  Scanning for consul KV")
		if err := c.processConsulKV(x.Filter("kv"), env, nil); err != nil {
			return err
		}

		c.Environments.Add(env)
	}

	return nil
}
