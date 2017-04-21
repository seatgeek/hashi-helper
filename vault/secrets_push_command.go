package vault

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/seatgeek/hashi-helper/config"
	helper "github.com/seatgeek/hashi-helper/vault/helper"
	cli "gopkg.in/urfave/cli.v1"
)

// PushSecretsCommand ...
func PushSecretsCommand(c *cli.Context) error {
	config, err := config.NewConfigFromDirectory(c.GlobalString("config-dir"))
	if err != nil {
		return err
	}

	var engine helper.SecretWriter

	if c.Bool("isolated") {
		env := c.GlobalString("environment")
		if env == "" {
			return fmt.Errorf("Isolated writer requires a environment value (--environment or ENV[ENVIRONMENT])")
		}

		if !config.Contains(env) {
			return fmt.Errorf("Could not find any environment with name %s in configuration", env)
		}

		engine = helper.IsolatedSecretWriter{}
		err := engine.WriteEnvironment(env, config[env])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		engine = helper.SharedSecretWriter{}

		for env, apps := range config {
			engine.WriteEnvironment(env, apps)
		}
	}

	return nil
}
