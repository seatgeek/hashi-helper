package vault

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/seatgeek/hashi-helper/config"
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
		return fmt.Errorf("Isolated writer requires a environment value (--environment or ENV[ENVIRONMENT])")
	}

	spew.Dump(config)

	// if !config.Contains(env) {
	// 	return fmt.Errorf("Could not find any environment with name %s in configuration", env)
	// }

	// var engine helper.SecretWriter

	// engine = helper.IsolatedSecretWriter{}
	// err := engine.WriteEnvironment(env, config[env])
	// if err != nil {
	// 	log.Fatal(err)
	// }

	return nil
}
