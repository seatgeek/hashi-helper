package config

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
)

// Application ...
type Application struct {
	Name        string
	Environment *Environment
}

// Equal ...
func (a *Application) Equal(o *Application) bool {
	if a.Name != o.Name {
		return false
	}

	if !a.Environment.Equal(o.Environment) {
		return false
	}

	return true
}

// Applications ...
type Applications []*Application

// Add ...
func (a *Applications) Add(application *Application) {
	if !a.Exists(application) {
		*a = append(*a, application)
	}
}

// Exists ...
func (a *Applications) Exists(application *Application) bool {
	for _, existing := range *a {
		if application.Equal(existing) {
			return true
		}
	}

	return false
}

// Get ...
func (a *Applications) Get(application *Application) *Application {
	for _, existing := range *a {
		if application.Equal(existing) {
			return existing
		}
	}

	return nil
}

// GetOrSet ...
func (a *Applications) GetOrSet(application *Application) *Application {
	existing := a.Get(application)
	if existing != nil {
		return existing
	}

	a.Add(application)
	return application
}

func (a *Applications) List() []string {
	res := []string{}

	for _, app := range *a {
		res = append(res, app.Name)
	}

	return res
}

func (c *Config) parseApplicationStanza(applicationsAST *ast.ObjectList, environment *Environment) error {
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

		if c.TargetApplication != "" && appName != c.TargetApplication {
			c.logger.Debugf("Skipping application %s (!= %s)", appName, c.TargetApplication)
			continue
		}

		// Check for valid keys inside an application stanza
		x := appAST.Val.(*ast.ObjectType).List
		valid := []string{"secret", "secrets", "policy", "kv"}
		if err := c.checkHCLKeys(x, valid); err != nil {
			return err
		}

		application := c.Applications.GetOrSet(&Application{Environment: environment, Name: appName})

		environment.Applications.Add(application)

		c.logger.Debug("Scanning for vault secret{}")
		if err := c.parseVaultSecretStanza(x.Filter("secret"), environment, application); err != nil {
			return err
		}

		c.logger.Debug("Scanning for vault secrets{}")
		if err := c.parseVaultSecretsStanza(x.Filter("secrets"), environment, application); err != nil {
			return err
		}

		c.logger.Debug("Scanning for policy{}")
		if err := c.parseVaultPolicyStanza(x.Filter("policy"), environment, application); err != nil {
			return err
		}

		c.logger.Debug("Scanning for kv{}")
		if err := c.parseConsulKVStanza(x.Filter("kv"), environment, application); err != nil {
			return err
		}

		c.Applications.Add(application)
	}

	return nil
}
