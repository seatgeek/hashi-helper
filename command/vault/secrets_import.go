package vault

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/seatgeek/hashi-helper/command/vault/helper"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
)

// SecretsImport ...
func SecretsImport(c *cli.Context) error {
	secrets := helper.IndexRemoteSecrets(c.GlobalString("environment"), c.GlobalInt("concurrency"))

	secrets, err := helper.ReadRemoteSecrets(secrets, c.GlobalInt("concurrency"))
	if err != nil {
		return err
	}

	output := make(map[string]map[string]string)

	for _, secret := range secrets {
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

		for k, v := range secret.VaultSecret.Data {
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

			dir := c.GlobalString("config-dir") + "/" + _env

			if _, err := os.Stat(dir); os.IsNotExist(err) {
				var mode os.FileMode
				mode = 0777
				os.Mkdir(dir, mode)
			}

			file := dir + "/app-" + _app + ".hcl"
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
