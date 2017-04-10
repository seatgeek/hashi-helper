package vault

import (
	log "github.com/Sirupsen/logrus"
	cfg "github.com/seatgeek/vault-restore/config"
	cli "gopkg.in/urfave/cli.v1"
)

// LocalSecretsListCommand ...
func LocalSecretsListCommand(c *cli.Context) error {
	config, err := cfg.NewConfigFromDirectory("./conf.d")
	if err != nil {
		return err
	}

	for envName, apps := range config {
		envLogger := log.WithFields(log.Fields{"env": envName})

		for appName, app := range apps {
			appLogger := envLogger.WithFields(log.Fields{"app": appName})

			for secretKey, secretValues := range app.Secrets {
				secretLogger := appLogger.WithFields(log.Fields{"secret": secretKey})

				for k, v := range secretValues {
					secretLogger.Printf("%s = %s", k, v)
				}
			}
		}

		log.Println()
	}

	return nil
}
