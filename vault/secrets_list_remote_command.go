package vault

import (
	log "github.com/Sirupsen/logrus"
	"github.com/seatgeek/hashi-helper/config"
	helper "github.com/seatgeek/hashi-helper/vault/helper"
	cli "gopkg.in/urfave/cli.v1"
)

// ListSecretsRemoteCommand ...
func ListSecretsRemoteCommand(c *cli.Context) error {
	secrets := helper.IndexRemoteSecrets(c.GlobalString("environment"))

	if c.Bool("detailed") {
		printDetailedSecrets(secrets)
		return nil
	}

	log.Println()
	for _, secret := range secrets {
		log.Infof("%s @ %s: %s", secret.Application, secret.Environment, secret.Path)
	}

	return nil
}

func printDetailedSecrets(paths config.Secrets) {
	secrets, err := helper.ReadRemoteSecrets(paths)
	if err != nil {
		log.Fatal(err)
	}

	for _, secret := range secrets {
		log.Println()
		log.Infof("%s @ %s: %s", secret.Application, secret.Environment, secret.Path)

		for k, v := range secret.Secret.Data {
			switch vv := v.(type) {
			case string:
				log.Info("  ⇛ ", k, " = ", vv)
			case int:
				log.Println("  ⇛ ", k, " = ", vv)
			default:
				log.Panic("  ⇛ ", k, "is of a type I don't know how to handle")
			}
		}
	}
}
