package config

import (
	"fmt"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"

	vault "github.com/hashicorp/vault/api"
)

// parseSecretStanza
// parse out `environment -> application -> secret {}` stanza
func parseSecretStanza(list *ast.ObjectList, envName, appName string) (Secrets, error) {
	secrets := make(Secrets)

	for _, secretData := range list.Items {
		if len(secretData.Keys) != 1 {
			return nil, fmt.Errorf("Missing secret name in line %+v", secretData.Keys[0].Pos())
		}

		secretName := secretData.Keys[0].Token.Value().(string)

		var m map[string]interface{}
		if err := hcl.DecodeObject(&m, secretData.Val); err != nil {
			return nil, err
		}

		secret := Secret{
			Path:        secretName,
			Environment: envName,
			Application: appName,
			Key:         secretName,
			Secret: &vault.Secret{
				Data: m,
			},
		}

		secrets[secretName] = secret
	}

	return secrets, nil
}
