package vault

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// AuthPush ...
func AuthPush(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	return AuthPushWithConfig(c, config)
}

// AuthPushWithConfig ...
func AuthPushWithConfig(c *cli.Context, config *config.Config) error {
	env := c.GlobalString("environment")
	if env == "" {
		return fmt.Errorf("Pushing auth backends require a 'environment' value (--environment or ENV[ENVIRONMENT])")
	}

	if !config.Environments.Contains(env) {
		return fmt.Errorf("Could not find any environment with name %s in configuration", env)
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	auths, err := client.Sys().ListAuth()

	if err != nil {
		return err
	}

	for _, auth := range config.VaultAuths {

		// Auth

		if _, ok := auths[auth.Name+"/"]; !ok {
			log.Printf("Creating auth backend %s", auth.Name)
			err := client.Sys().EnableAuth(auth.Name, auth.Type, auth.Description)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Printf("Auth backend %s already exist", auth.Name)
		}

		// Auth config

		for _, config := range auth.Config {
			configPath := fmt.Sprintf("auth/%s/config/%s", auth.Name, config.Name)
			log.Printf("  Writing auth backend config: %s", configPath)

			s, err := client.Logical().Write(configPath, config.Data)
			if err != nil {
				log.Fatal(err)
			}

			printRemoteSecretWarnings(s)
		}

		// Auth roles

		for _, role := range auth.Roles {
			rolePath := fmt.Sprintf("auth/%s/role/%s", auth.Name, role.Name)
			log.Printf("  Writing auth backend role: %s", rolePath)

			s, err := client.Logical().Write(rolePath, role.Data)
			if err != nil {
				log.Fatal(err)
			}

			printRemoteSecretWarnings(s)
		}
	}

	return nil
}
