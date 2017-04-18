package main

import (
	"os"
	"runtime"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/seatgeek/hashi-helper/config"
	"github.com/seatgeek/hashi-helper/vault"
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
			Value:       runtime.NumCPU() * 2,
			Usage:       "How many parallel requests to run in parallel against remote servers (2 * CPU Cores)",
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
			Name:  "vault-local-list-secrets",
			Usage: "Print a list of local secrets",
			Action: func(c *cli.Context) error {
				return vault.LocalSecretsListCommand(c)
			},
		},
		{
			Name:  "vault-local-write-secrets",
			Usage: "Write remote secrets to local disk",
			Action: func(c *cli.Context) error {
				return vault.LocalSecretsWriteCommand(c)
			},
		},
		{
			Name:  "vault-remote-list-secrets",
			Usage: "Print a list of remote secrets",
			Action: func(c *cli.Context) error {
				return vault.RemoteSecretsListCommand(c)
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "detailed",
					Usage:  "Only show keys, or also expand and show the secret values (highly sensitive!)",
					EnvVar: "DETAILED",
				},
			},
		},
		{
			Name:  "vault-remote-write-secrets",
			Usage: "Write local secrets to remote Vault instance",
			Action: func(c *cli.Context) error {
				return vault.RemoteSecretsWriteCommand(c)
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "isolated",
					Usage:  "Write to the cluster as if its isolated env (e.g. don't encode environment into the path)",
					EnvVar: "isolated",
				},
			},
		},
		{
			Name:  "vault-remote-write-policies",
			Usage: "Write application read-only policies to remote Vault instance",
			Action: func(c *cli.Context) error {
				return vault.RemotePoliciesWriteCommand(c)
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "isolated",
					Usage:  "Write to the cluster as if its isolated env (e.g. don't encode environment into the path)",
					EnvVar: "isolated",
				},
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
