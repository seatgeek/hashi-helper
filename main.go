package main

import (
	"os"
	"runtime"
	"sort"

	log "github.com/Sirupsen/logrus"
	allCommand "github.com/seatgeek/hashi-helper/command"
	consulCommand "github.com/seatgeek/hashi-helper/command/consul"
	vaultCommand "github.com/seatgeek/hashi-helper/command/vault"
	"github.com/seatgeek/hashi-helper/config"
	"github.com/seatgeek/hashi-helper/command"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "hashi-helper"
	app.Usage = "Manage cleartext and encrypted secrets, configuration, and more for Vault and Consul."
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
		cli.BoolFlag{
		  Name:        "warn-unencrypted",
		  Usage:       "Issue a warning if unencrypted HCL files are discovered",
		  EnvVar:      "WARN_UNENCRYPTED",
		  Destination: &config.WarnUnencrypted,
		},
		cli.BoolFlag{
			Name:        "no-keybase-team",
			Usage:       "Do not use the keybase team name in config file (if specified) to re-sign encrypted files post-edit",
			EnvVar:      "NO_KEYBASE_TEAM",
			Destination: &command.NoKeybaseTeam,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "push-all",
			Usage: "Push all consul and vault settings",
			Action: func(c *cli.Context) error {
				return allCommand.PushAll(c)
			},
		},
		{
			Name:  "edit-file",
			Usage: "Edit a Keybase PGP-encrypted file (config file or otherwise)",
			Action: func(c *cli.Context) error {
				return allCommand.EditEncryptedFile(c)
			},
		},
		{
			Name:  "vault-profile-use",
			Usage: "Change your current vault env profile",
			Action: func(c *cli.Context) error {
				return vaultCommand.UseProfile(c)
			},
		},
		{
			Name:  "vault-profile-edit",
			Usage: "Edit your current vault env profile",
			Action: func(c *cli.Context) error {
				return vaultCommand.EditProfile(c)
			},
		},
		{
			Name:  "vault-unseal-keybase",
			Usage: "Unseal Vault with keybase encrypted unseal tokens",
			Action: func(c *cli.Context) error {
				return vaultCommand.UnsealKeybase(c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "unseal-key",
					Usage:  "The raw base64 encoded and encrypted unseal key",
					EnvVar: "VAULT_UNSEAL_KEY",
				},
				cli.StringFlag{
					Name:   "vault-protocol",
					Usage:  "The protocol to use when talking to vault (http or https)",
					EnvVar: "VAULT_PROTOCOL",
					Value:  "http",
				},
				cli.StringFlag{
					Name:   "consul-service-name",
					Usage:  "A consul service name to find vault instances from",
					EnvVar: "CONSUL_SERVICE_NAME",
				},
				cli.StringFlag{
					Name:   "consul-service-tag",
					Usage:  "A consul tag name to filter found consul services by",
					EnvVar: "CONSUL_SERVICE_TAG",
					Value:  "standby",
				},
			},
		},
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
			Name:  "vault-push-all",
			Usage: "Push all known resources to remote Vault",
			Action: func(c *cli.Context) error {
				return vaultCommand.PushAll(c)
			},
		},
		{
			Name:  "vault-push-secrets",
			Usage: "Write local secrets to remote Vault",
			Action: func(c *cli.Context) error {
				return vaultCommand.SecretsPush(c)
			},
		},
		{
			Name:  "vault-delete-secrets",
			Usage: "Delete remote secrets from Vault, leaves local secrets unchanged",
			Action: func(c *cli.Context) error {
				return vaultCommand.SecretsDelete(c)
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "skip-confirm",
					Usage:  "Do NOT ask for confirmation at each remote Vault prefix delete",
				},
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
		{
			Name:  "vault-push-auth",
			Usage: "Write vault auth backends to remote Vault instance",
			Action: func(c *cli.Context) error {
				return vaultCommand.AuthPush(c)
			},
		},
		{
			Name:  "vault-create-token",
			Usage: "Create a vault token",
			Action: func(c *cli.Context) error {
				return vaultCommand.CreateToken(c)
			},
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name: "keybase",
				},
				cli.StringFlag{
					Name: "id",
				},
				cli.StringFlag{
					Name: "display-name",
				},
				cli.StringFlag{
					Name: "ttl",
				},
				cli.StringFlag{
					Name: "period",
				},
				cli.BoolFlag{
					Name: "orphan",
				},
				cli.StringSliceFlag{
					Name: "policy",
				},
			},
		},
		{
			Name:  "vault-find-token",
			Usage: "Find vault token matching a name",
			Action: func(c *cli.Context) error {
				return vaultCommand.FindToken(c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "filter-name",
					Usage: "Only match tokens that has this fuzzy name in their display name",
				},
				cli.StringFlag{
					Name:  "filter-policy",
					Usage: "Only match tokens that have this policy",
				},
				cli.StringFlag{
					Name:  "filter-path",
					Usage: "Only match tokens that was created from this path",
				},
				cli.StringFlag{
					Name:  "filter-meta-username",
					Usage: "Only match tokens from matching meta[username] (e.g. from GitHub auth backend)",
				},
				cli.BoolFlag{
					Name:  "filter-orphan",
					Usage: "Only match tokens that are orphans",
				},
				cli.BoolFlag{
					Name:  "delete-matches",
					Usage: "Delete all tokens that match the filters",
				},
			},
		},
		{
			Name:  "consul-push-all",
			Usage: "Push all known consul configs to remote Consul cluster",
			Action: func(c *cli.Context) error {
				return consulCommand.PushAll(c)
			},
		},
		{
			Name:  "consul-push-services",
			Usage: "Push all known consul services to remote Consul cluster",
			Action: func(c *cli.Context) error {
				return consulCommand.ServicesPush(c)
			},
		},
		{
			Name:  "consul-push-kv",
			Usage: "Push all known consul kv to remote Consul cluster",
			Action: func(c *cli.Context) error {
				return consulCommand.KVPush(c)
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
