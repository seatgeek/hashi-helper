package vault

import (
	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// SecretsListLocal ...
func SecretsListLocal(c *cli.Context) error {
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
