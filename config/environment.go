package config

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
)

// Environment struct
type Environment struct {
	Name         string
	Applications Applications
}

// equal ...
func (e *Environment) equal(o *Environment) bool {
	return e.Name == o.Name
}

// Environments struct
type Environments []*Environment

// add ...
func (e *Environments) add(environment *Environment) {
	if !e.exists(environment) {
		*e = append(*e, environment)
	}
}

// exists ...
func (e *Environments) exists(environment *Environment) bool {
	for _, existing := range *e {
		if environment.equal(existing) {
			return true
		}
	}

	return false
}

// Containts ...
func (e *Environments) Contains(environmentName string) bool {
	for _, existing := range *e {
		if existing.Name == environmentName {
			return true
		}
	}

	return false
}

// get ...
func (e *Environments) get(environment *Environment) *Environment {
	for _, existing := range *e {
		if environment.equal(existing) {
			return existing
		}
	}

	return nil
}

// getOrSet ...
func (e *Environments) getOrSet(environment *Environment) *Environment {
	existing := e.get(environment)
	if existing != nil {
		return existing
	}

	e.add(environment)
	return environment
}

func (e *Environments) list() []string {
	res := []string{}

	for _, env := range *e {
		res = append(res, env.Name)
	}

	return res
}

// Parse root environment stanza
func (c *Config) processEnvironments(list *ast.ObjectList) error {
	valid := []string{"environment"}
	if err := c.checkHCLKeys(list, valid); err != nil {
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
			if c.shouldSkipEnvironment(envName, c.targetEnvironment) {
				c.logger.Debugf("  Skipping environment %s (%s != %s)", envName, envName, c.targetEnvironment)
				continue
			}

			// Rewrite "*" to the actual target environment
			if envName == "*" {
				envName = c.targetEnvironment
			}

			// check for valid keys inside an environment stanza
			x := envAST.Val.(*ast.ObjectType).List
			valid := []string{"application", "auth", "policy", "mount", "secret", "secrets", "service", "kv"}
			if err := c.checkHCLKeys(x, valid); err != nil {
				return err
			}

			env := c.Environments.getOrSet(&Environment{Name: envName})

			c.logger.Debug("Scanning for application{}")
			if err := c.parseApplicationStanza(x.Filter("application"), env); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault auth{}")
			if err := c.parseVaultAuthStanza(x.Filter("auth"), env); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault secret{}")
			if err := c.parseVaultSecretStanza(x.Filter("secret"), env, nil); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault secrets{}")
			if err := c.parseVaultSecretsStanza(x.Filter("secrets"), env, nil); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault policy{}")
			if err := c.parseVaultPolicyStanza(x.Filter("policy"), env, nil); err != nil {
				return err
			}

			c.logger.Debug("Scanning for vault mount{}")
			if err := c.parseVaultMountStanza(x.Filter("mount"), env); err != nil {
				return err
			}

			c.logger.Debug("Scanning for consul service{}")
			if err := c.parseConsulServiceStanza(x.Filter("service"), env); err != nil {
				return err
			}

			c.logger.Debug("Scanning for consul kv{}")
			if err := c.parseConsulKVStanza(x.Filter("kv"), env, nil); err != nil {
				return err
			}

			c.Environments.add(env)
		}

		// spew.Dump(c)
	}

	return nil
}

func (c *Config) shouldSkipEnvironment(parsedEnv, targetEnv string) bool {
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
