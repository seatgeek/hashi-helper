package command

import (
	consul "github.com/seatgeek/hashi-helper/command/consul"
	vault "github.com/seatgeek/hashi-helper/command/vault"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// PushAll ...
func PushAll(cli *cli.Context) error {
	config, err := config.NewConfigFromCLI(cli)
	if err != nil {
		return err
	}

	// Vault
	if err := vault.PushAllWithConfig(cli, config); err != nil {
		return err
	}

	// Consul
	return consul.PushAllWithConfig(cli, config)
}
