package main

import (
	"os"
	"runtime"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	api "github.com/hashicorp/vault/api"
	cli "gopkg.in/urfave/cli.v1"
)

var defaultConcurrency int

func main() {
	app := cli.NewApp()
	app.Name = "vault-manager"
	app.Usage = "easily restore / snapshot your secrets"
	app.Flags = []cli.Flag{
		// cli.BoolFlag{
		// 	Name:   "lint",
		// 	Usage:  "Don't run any Add, Update or Delete operations against Vault",
		// 	EnvVar: "LINT",
		// },
		cli.IntFlag{
			Name:        "concurrency",
			Value:       runtime.NumCPU() * 2,
			Usage:       "How many parallel requests to run in parallel against remote servers (2 * CPU Cores)",
			EnvVar:      "CONCURRENCY",
			Destination: &defaultConcurrency,
		},
		cli.StringFlag{
			Name:   "log-level",
			Value:  "info",
			Usage:  "Debug level (debug, info, warn/warning, error, fatal, panic)",
			EnvVar: "LOG_LEVEL",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "list-local",
			Usage: "Print a list of local secrets",
			Action: func(c *cli.Context) error {
				return listLocalSecretsCommand(c)
			},
		},
		{
			Name:  "list-remote",
			Usage: "Print a list of remote secrets",
			Action: func(c *cli.Context) error {
				return listRemoteSecretsCommand(c)
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
			Name:  "write-remote",
			Usage: "Write local secrets to remote Vault instance",
			Action: func(c *cli.Context) error {
				return writeRemoteSecretsCommand(c)
			},
		},
		{
			Name:  "clean-remote",
			Usage: "Delete remote Vault secrets not in the local catalog",
			Action: func(c *cli.Context) error {
				return deleteRemoteSecretsCommand(c)
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

func merp() {
	client, err := api.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	l, err := client.Logical().List("secret/")
	if err != nil {
		log.Fatal(err)
	}

	spew.Dump(l.Data["keys"])

	m := make(map[string]interface{})
	m["hello"] = "world"

	s, err := client.Logical().Write("secret/cw/test", m)
	if err != nil {
		log.Fatal(err)
	}

	spew.Dump(s)

	// spew.Dump(config)
	// spew.Dump(err)
}
