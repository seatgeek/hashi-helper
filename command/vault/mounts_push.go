package vault

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// MountsPush ...
func MountsPush(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	return MountsPushWithConfig(c, config)
}

// MountsPushWithConfig ...
func MountsPushWithConfig(c *cli.Context, config *config.Config) error {
	env := c.GlobalString("environment")
	if env == "" {
		return fmt.Errorf("Pushing mounts require a 'environment' value (--environment or ENV[ENVIRONMENT])")
	}

	if !config.Environments.Contains(env) {
		return fmt.Errorf("Could not find any environment with name %s in configuration", env)
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return err
	}

	for _, mount := range config.VaultMounts {

		// MOUNT POINT

		mountLogicalName := mount.Name + "/"
		if _, ok := mounts[mountLogicalName]; !ok {
			
			err := client.Sys().Mount(mount.Name, mount.MountInput())
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Printf("Mount %s already exist", mountLogicalName)
		}

		// MOUNT CONFIG

		for _, config := range mount.Config {
			configPath := fmt.Sprintf("%s/config/%s", mount.Name, config.Name)

			log.Printf("  Writing mount config %s", configPath)

			s, err := client.Logical().Write(configPath, config.Data)
			if err != nil {
				log.Fatal(err)
			}

			printRemoteSecretWarnings(s)
		}

		// MOUNT ROLES

		for _, role := range mount.Roles {
			rolePath := fmt.Sprintf("%s/roles/%s", mount.Name, role.Name)

			log.Printf("  Writing role %s", rolePath)

			s, err := client.Logical().Write(rolePath, role.Data)
			if err != nil {
				log.Fatal(err)
			}

			printRemoteSecretWarnings(s)
		}
	}

	return nil
}

func printRemoteSecretWarnings(s *api.Secret) {
	if s != nil && len(s.Warnings) > 0 {
		for _, warn := range s.Warnings {
			log.Warn("    REMOTE WARNING: " + warn)
		}
	}
}
