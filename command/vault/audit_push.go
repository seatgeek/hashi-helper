package vault

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
)

// AuditPush ...
func AuditPush(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	return AuditPushWithConfig(c, config)
}

// AuditPushWithConfig ...
func AuditPushWithConfig(c *cli.Context, config *config.Config) error {
	log.Info("Pushing Vault Audit")

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

	audits, err := client.Sys().ListAudit()
	if err != nil {
		return err
	}

	for _, audit := range config.VaultAudits {
		log.Printf("  Writing audit path: %s", audit.Path)

		path := fmt.Sprintf("/sys/audit/%s", audit.Path)
		if _, ok := audits[audit.Path+"/"]; ok {
			s, err := client.Logical().Delete(path)
			if err != nil {
				log.Fatal(err)
			}

			// Give Vault a little bit of time to complete the DELETE operation above
			time.Sleep(1 * time.Second)

			printRemoteSecretWarnings(s)
		}

		s, err := client.Logical().Write(path, audit.ToMap())
		if err != nil {
			log.Fatal(err)
		}

		printRemoteSecretWarnings(s)
	}

	return nil
}
