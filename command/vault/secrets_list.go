package vault

import (
	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/seatgeek/hashi-helper/command/vault/helper"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// SecretsList ...
func SecretsList(c *cli.Context) error {
	if c.GlobalBool("remote") {
		return secretListRemote(c)
	}

	return secretListLocal(c)
}

func secretListLocal(c *cli.Context) error {
	config, err := config.NewConfig(c.GlobalString("config-dir"))
	if err != nil {
		return err
	}

	for _, secret := range config.Secrets {
		logger := log.WithFields(log.Fields{
			"env":    secret.Environment.Name,
			"app":    secret.Application.Name,
			"secret": secret.Key,
		})

		for k, v := range secret.Secret.Data {
			logger.Printf("%s = %s", k, v)
		}

		log.Println()
	}

	spew.Dump(config)

	return nil
}

// secretListRemote ...
func secretListRemote(c *cli.Context) error {
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
