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
	config, err := config.NewConfig(c.GlobalString("config-dir"))
	if err != nil {
		return err
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	env := c.GlobalString("environment")
	if env == "" {
		return fmt.Errorf("Isolated writer requires a environment value (--environment or ENV[ENVIRONMENT])")
	}

	// if !config.Contains(env) {
	// 	return fmt.Errorf("Could not find any environment with name %s in configuration", env)
	// }

	for _, policy := range config.Policies {
		var policyPath, policyName string

		policyPath = fmt.Sprintf("secret/%s/*", policy.Application.Name)
		policyName = fmt.Sprintf("app-%s-readonly", policy.Application.Name)

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

		log.Println()
	}

	return nil
}
