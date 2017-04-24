package vault

import (
	"fmt"
	"log"

	"github.com/seatgeek/hashi-helper/config"
	"github.com/seatgeek/hashi-helper/vault/helper"
	cli "gopkg.in/urfave/cli.v1"
)

// PushSecretsCommand ...
func PushSecretsCommand(c *cli.Context) error {
	config, err := config.NewConfig(c.GlobalString("config-dir"))
	if err != nil {
		return err
	}

	env := c.GlobalString("environment")
	if env == "" {
		return fmt.Errorf("Secret writer requires a environment value (--environment or ENV[ENVIRONMENT])")
	}

	if !config.Environments.Contains(env) {
		return fmt.Errorf("Could not find any environment with name %s in configuration", env)
	}

	engine := helper.SecretWriter{}
	for _, secret := range config.Secrets {
		err := engine.WriteSecret(secret)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
