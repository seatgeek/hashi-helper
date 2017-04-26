package main

import (
	"os"
	"runtime"
	"sort"

	log "github.com/Sirupsen/logrus"
	vaultCommand "github.com/seatgeek/hashi-helper/command/vault"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "vault-manager"
	app.Usage = "easily restore / snapshot your secrets"
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "concurrency",
			Value:       runtime.NumCPU() * 3,
			Usage:       "How many parallel requests to run in parallel against remote servers (3 * CPU Cores)",
			EnvVar:      "CONCURRENCY",
			Destination: &config.DefaultConcurrency,
		},
		cli.StringFlag{
			Name:   "log-level",
			Value:  "info",
			Usage:  "Debug level (debug, info, warn/warning, error, fatal, panic)",
			EnvVar: "LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "config-dir",
			Value:  "./conf.d",
			Usage:  "Config directory to read and write from",
			EnvVar: "CONFIG_DIR",
		},
		cli.StringFlag{
			Name:   "config-file",
			Value:  "",
			Usage:  "Config file to read from, if you don't want to scan a directory recursively",
			EnvVar: "CONFIG_FILE",
		},
		cli.StringFlag{
			Name:        "environment",
			Usage:       "The environment to process for (default: all env)",
			EnvVar:      "ENVIRONMENT",
			Destination: &config.TargetEnvironment,
		},
		cli.StringFlag{
			Name:        "application",
			Usage:       "The application to process for (default: all applications)",
			EnvVar:      "APPLICATION",
			Destination: &config.TargetApplication,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "vault-list-secrets",
			Usage: "Print a list of local or remote secrets",
			Action: func(c *cli.Context) error {
				return vaultCommand.SecretsList(c)

			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "remote",
					Usage:  "Target remote Vault instance instead of local secret stash",
					EnvVar: "REMOTE",
				},
				cli.BoolFlag{
					Name:   "detailed",
					Usage:  "Only show keys, or also expand and show the secret values (highly sensitive!)",
					EnvVar: "DETAILED",
				},
			},
		},
		{
			Name:  "vault-pull-secrets",
			Usage: "Write remote secrets to local disk",
			Action: func(c *cli.Context) error {
				return vaultCommand.SecretsPull(c)
			},
		},
		{
			Name:  "vault-import-secrets",
			Usage: "Write remote secrets to local disk (legacy)",
			Action: func(c *cli.Context) error {
				return vaultCommand.SecretsImport(c)
			},
		},
		{
			Name:  "vault-push-secrets",
			Usage: "Write local secrets to remote Vault instance",
			Action: func(c *cli.Context) error {
				return vaultCommand.SecretsPush(c)
			},
		},
		{
			Name:  "vault-push-policies",
			Usage: "Write application read-only policies to remote Vault instance",
			Action: func(c *cli.Context) error {
				return vaultCommand.PoliciesPush(c)
			},
		},
		{
			Name:  "vault-push-mounts",
			Usage: "Write vault mounts to remote Vault instance",
			Action: func(c *cli.Context) error {
				return vaultCommand.MountsPush(c)
			},
		},
	}
	app.Before = func(c *cli.Context) error {
		// convert the human passed log level into logrus levels
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatal(err)
		}
		log.SetLevel(level)

		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	app.Run(os.Args)
}
