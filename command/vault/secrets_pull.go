package vault

import (
	"fmt"
	"io/ioutil"
	"os"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/seatgeek/hashi-helper/command/vault/helper"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// SecretsPull ...
func SecretsPull(c *cli.Context) error {
	log.Fatal("NOT IMPLEMENTED YET")

	env := c.GlobalString("environment")
	if env == "" {
		return fmt.Errorf("Pulling secrets require a 'environment' value (--environment or ENV[ENVIRONMENT])")
	}

	config, err := config.NewConfig(c.GlobalString("config-dir"))
	if err != nil {
		return err
	}

	if !config.Environments.Contains(env) {
		return fmt.Errorf("Could not find any environment with name %s in configuration - did you mean `vault-import-secrets`?", env)
	}

	remoteSecrets := helper.IndexRemoteSecrets(c.GlobalString("environment"))
	remoteSecrets, err = helper.ReadRemoteSecrets(remoteSecrets)
	if err != nil {
		return err
	}

	output := make(map[string]map[string]string)

	for _, secret := range remoteSecrets {
		env := secret.Environment.Name
		app := secret.Application.Name

		if _, ok := output[env]; !ok {
			output[env] = make(map[string]string)
		}

		if _, ok := output[env][app]; !ok {
			output[env][app] = `policy "{app}-read-only" {
				path "secret/{app}/*" {
					capabilities = ["read", "list"]
				}
			}`

			output[env][app] = strings.Replace(output[env][app], "{app}", app, -1)
		}

		output[env][app] = output[env][app] + fmt.Sprintf("\t\tsecret \"%s\" {\n", secret.Key)

		for k, v := range secret.Secret.Data {
			switch vv := v.(type) {
			case string:
				output[env][app] = output[env][app] + fmt.Sprintf("\t\t\t%s = \"%s\"\n", k, helper.EscapeValue(vv))
			case int:
				output[env][app] = output[env][app] + fmt.Sprintf("\t\t\t%s = %d\n", k, vv)
			default:
				log.Panic("  ⇛ ", k, "is of a type I don't know how to handle")
			}
		}

		output[env][app] = output[env][app] + "\t\t}\n"
	}

	for _env, _apps := range output {
		for _app, _output := range _apps {
			content := fmt.Sprintf("environment \"%s\" {\n", _env)
			content = content + fmt.Sprintf("\tapplication \"%s\" {\n", _app)
			content = content + _output
			content = content + fmt.Sprintf("\t}\n")

			content = content + fmt.Sprintf("}\n\n")

			dir := c.GlobalString("config-dir") + "/apps/"

			if _, err := os.Stat(dir); os.IsNotExist(err) {
				var mode os.FileMode
				mode = 0777
				os.MkdirAll(dir, mode)
			}

			file := dir + _app + ".hcl"
			err := ioutil.WriteFile(file, []byte(content), 0644)
			if err != nil {
				return err
			}
			log.Infof("Wrote file: %s", file)

			output, err := helper.FormatHclFile(file)
			if err != nil {
				return err
			}

			log.Infof("Formatted file: %s (%s)", file, output.String())
		}
	}

	return nil
}
