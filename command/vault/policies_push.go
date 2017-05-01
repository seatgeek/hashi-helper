package vault

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// PoliciesPush ...
func PoliciesPush(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	return PoliciesPushWithConfig(c, config)
}

// PoliciesPushWithConfig ...
func PoliciesPushWithConfig(c *cli.Context, config *config.Config) error {
	env := c.GlobalString("environment")
	if env == "" {
		return fmt.Errorf("Pushing policies require a 'environment' value (--environment or ENV[ENVIRONMENT])")
	}

	if !config.Environments.Contains(env) {
		return fmt.Errorf("Could not find any environment with name %s in configuration", env)
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	for _, policy := range config.VaultPolicies {
		var policyName, policyContent string

		policyName = policy.Name
		policyContent = policy.Raw

		log.Printf("Writing policy %s", policyName)
		log.Debugf("  content: %s", policyContent)

		err := client.Sys().PutPolicy(policyName, policyContent)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
