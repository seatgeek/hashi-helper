package main

import (
	"os"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	api "github.com/hashicorp/vault/api"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "vault-manager"
	app.Usage = "easily restore / snapshot your secrets"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "lint",
			Usage: "Don't run any Add, Update or Delete operations against Vault",
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
