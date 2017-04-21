package vault

import (
	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// ListSecretsLocalCommand ...
func ListSecretsLocalCommand(c *cli.Context) error {
	config, err := config.NewConfigFromDirectory(c.GlobalString("config-dir"))
	if err != nil {
		return err
	}

	for envName, env := range config {
		envLogger := log.WithFields(log.Fields{"env": envName})

		for appName, app := range env.Applications {
			appLogger := envLogger.WithFields(log.Fields{"app": appName})

			for secretKey, secretValues := range app.Secrets {
				secretLogger := appLogger.WithFields(log.Fields{"secret": secretKey})

				for k, v := range secretValues.Secret.Data {
					secretLogger.Printf("%s = %s", k, v)
				}
			}
		}

		log.Println()
	}

	spew.Dump(config)

	return nil
}
