package config

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	vault "github.com/hashicorp/vault/api"
)

// parseSecretStanza
// parse out `environment -> application -> secret {}` stanza
func (c *Config) processVaultSecret(list *ast.ObjectList, env *Environment, app *Application) error {
	if len(list.Items) == 0 {
		return nil
	}

	c.logger = c.logger.WithField("stanza", "secrets")
	c.logger.Debugf("Found %d secrets{}", len(list.Items))
	for _, secretData := range list.Items {
		if len(secretData.Keys) != 1 {
			return fmt.Errorf("Missing secret name in line %+v", secretData.Keys[0].Pos())
		}

		var m map[string]interface{}
		if err := hcl.DecodeObject(&m, secretData.Val); err != nil {
			return err
		}

		secretName := secretData.Keys[0].Token.Value().(string)

		secret := &Secret{
			Application: app,
			Environment: env,
			Path:        secretName,
			Key:         secretName,
			Secret: &vault.Secret{
				Data: m,
			},
		}

		if c.VaultSecrets.Add(secret) == false {
			if secret.Application != nil {
				c.logger.Warnf("      Ignored duplicate secret '%s' -> '%s' -> '%s' in line %s", secret.Environment.Name, secret.Application.Name, secret.Key, secretData.Keys[0].Token.Pos)
			} else {
				c.logger.Warnf("      Ignored duplicate secret '%s' -> '%s' in line %s", secret.Environment.Name, secret.Key, secretData.Keys[0].Token.Pos)
			}
		}
	}

	return nil
}
func (c *Config) processVaultSecrets(list *ast.ObjectList, env *Environment, app *Application) error {
	if len(list.Items) == 0 {
		return nil
	}

	c.logger.Debugf("Found %d secrets{}", len(list.Items))
	for _, secretData := range list.Items {
		if len(secretData.Keys) != 0 {
			return errors.New("secrets{} stanza must not be named")
		}

		var m map[string]string
		if err := hcl.DecodeObject(&m, secretData.Val); err != nil {
			return err
		}

		for k, v := range m {
			secret := &Secret{
				Application: app,
				Environment: env,
				Path:        k,
				Key:         k,
				Secret: &vault.Secret{
					Data: map[string]interface{}{"value": v},
				},
			}

			if c.VaultSecrets.Add(secret) == false {
				if secret.Application != nil {
					c.logger.Warnf("Ignored duplicate secret '%s' -> '%s' -> '%s' in line %s", secret.Environment.Name, secret.Application.Name, secret.Key, secretData.Pos())
				} else {
					c.logger.Warnf("Ignored duplicate secret '%s' -> '%s' in line %s", secret.Environment.Name, secret.Key, secretData.Pos())
				}
			}
		}
	}

	return nil
}
