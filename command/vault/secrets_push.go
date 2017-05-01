package vault

import (
	"fmt"
	"log"

	"github.com/seatgeek/hashi-helper/command/vault/helper"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// SecretsPush ...
func SecretsPush(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	return SecretsPushWithConfig(c, config)
}

// SecretsPushWithConfig ...
func SecretsPushWithConfig(c *cli.Context, config *config.Config) error {
	env := c.GlobalString("environment")
	if env == "" {
		return fmt.Errorf("Secret writer requires a environment value (--environment or ENV[ENVIRONMENT])")
	}

	if !config.Environments.Contains(env) {
		return fmt.Errorf("Could not find any environment with name %s in configuration", env)
	}

	engine := helper.SecretWriter{}
	for _, secret := range config.VaultSecrets {
		err := engine.WriteSecret(secret)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
