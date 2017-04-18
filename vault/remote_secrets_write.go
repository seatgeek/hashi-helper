package vault

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// RemoteSecretsWriteCommand ...
func RemoteSecretsWriteCommand(c *cli.Context) error {
	config, err := config.NewConfigFromDirectory(c.GlobalString("config-dir"))
	if err != nil {
		return err
	}

	var engine SecretWriter

	if c.Bool("isolated") {
		env := c.GlobalString("environment")
		if env == "" {
			return fmt.Errorf("Isolated writer requires a environment value (--environment or ENV[ENVIRONMENT])")
		}

		if !config.Contains(env) {
			return fmt.Errorf("Could not find any environment with name %s in configuration", env)
		}

		engine = IsolatedSecretWriter{}
		err := engine.writeEnvironment(env, config[env])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		engine = SharedSecretWriter{}

		for env, apps := range config {
			engine.writeEnvironment(env, apps)
		}
	}

	return nil
}
