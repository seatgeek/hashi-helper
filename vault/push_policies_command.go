package vault

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// PushPoliciesCommand ...
func PushPoliciesCommand(c *cli.Context) error {
	config, err := config.NewConfigFromDirectory(c.GlobalString("config-dir"))
	if err != nil {
		return err
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	if c.Bool("isolated") {
		env := c.GlobalString("environment")
		if env == "" {
			return fmt.Errorf("Isolated writer requires a environment value (--environment or ENV[ENVIRONMENT])")
		}

		if !config.Contains(env) {
			return fmt.Errorf("Could not find any environment with name %s in configuration", env)
		}
	}

	for envName, env := range config {
		for appName := range env.Applications {
			var policyPath, policyName string

			if c.Bool("isolated") {
				policyPath = fmt.Sprintf("secret/%s/*", appName)
				policyName = fmt.Sprintf("app-%s-readonly", appName)
			} else {
				policyPath = fmt.Sprintf("secret/%s/%s/*", envName, appName)
				policyName = fmt.Sprintf("app-%s-%s-readonly", envName, appName)
			}

			policyContent := fmt.Sprintf(`
# created by hashi-helper
path "%s" {
	capabilities = ["read", "list"]
}
`, policyPath)

			log.Printf("policyName: %s", policyName)
			log.Printf("policyPath: %s", policyPath)
			log.Printf("policyContent: %s", policyContent)

			err := client.Sys().PutPolicy(policyName, policyContent)
			if err != nil {
				return err
			}
		}

		log.Println()
	}

	return nil
}
