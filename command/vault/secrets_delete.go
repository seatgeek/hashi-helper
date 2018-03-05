package vault

import (
	"fmt"
	"os"
	"strings"
	"bufio"

	log "github.com/Sirupsen/logrus"

	"github.com/seatgeek/hashi-helper/command/vault/helper"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// SecretsDelete ...
func SecretsDelete(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	return SecretsDeleteWithConfig(c, config)
}

// SecretsDeleteWithConfig ...
func SecretsDeleteWithConfig(c *cli.Context, config *config.Config) error {
	env := c.GlobalString("environment")
	if env == "" {
		return fmt.Errorf("Secret deleter requires a environment value (--environment or ENV[ENVIRONMENT])")
	}

	if !config.Environments.Contains(env) {
		return fmt.Errorf("Could not find any environment with name %s in configuration", env)
	}

	engine := helper.SecretDeleter{}
	for _, secret := range config.VaultSecrets {
		// optionally verify each delete
		do_delete := true
		if !c.Bool("skip-confirm") {
			log.Warnf("About to delete remote secret: secret/%s/%s/%s", env, secret.Application.Name, secret.Path)
			do_delete = confirmDelete("Continue?", 3)
		}

		if do_delete {
			err := engine.DeleteSecret(secret, env)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("           Skipping ...\n")
		}
	}

	return nil
}

func confirmDelete(s string, tries int) bool {
	r := bufio.NewReader(os.Stdin)
	for ; tries > 0; tries-- {
		fmt.Printf("           %s [y/n]: ", s)
		res, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		// Empty input (i.e. "\n")
		if len(res) < 2 {
			continue
		}
		return strings.ToLower(strings.TrimSpace(res))[0] == 'y'
	}

	return false
}
