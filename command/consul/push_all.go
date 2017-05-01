package consul

import (
	cfg "github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// PushAll ...
func PushAll(cli *cli.Context) error {
	config, err := cfg.NewConfigFromCLI(cli)
	if err != nil {
		return err
	}

	return PushAllWithConfig(cli, config)
}

// PushAllWithConfig ...
func PushAllWithConfig(cli *cli.Context, config *cfg.Config) error {
	if err := ServicesPushWithConfig(cli, config); err != nil {
		return err
	}

	return nil
}
