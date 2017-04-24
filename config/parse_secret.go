package config

import (
	"fmt"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"

	log "github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

// parseSecretStanza
// parse out `environment -> application -> secret {}` stanza
func (c *Config) processSecrets(list *ast.ObjectList, app *Application) error {
	if len(list.Items) == 0 {
		return nil
	}

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
			Environment: app.Environment,
			Path:        secretName,
			Key:         secretName,
			Secret: &vault.Secret{
				Data: m,
			},
		}

		if c.Secrets.Add(secret) == false {
			log.Warnf("      Ignored duplicate secret '%s' -> '%s' -> '%s' in line %s", secret.Environment.Name, secret.Application.Name, secret.Key, secretData.Keys[0].Token.Pos)
		}
	}

	return nil
}
