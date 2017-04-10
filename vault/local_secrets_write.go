package vault

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"

	"strings"

	log "github.com/Sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
)

func LocalSecretsWriteCommand(c *cli.Context) error {
	secrets := indexRemoteSecrets(c.GlobalString("environment"))

	secrets, err := readRemoteSecrets(secrets)
	if err != nil {
		return err
	}
	secrets.Sort()

	output := make(map[string]map[string]string)

	for _, secret := range secrets {
		env := secret.Environment
		app := secret.Application

		if _, ok := output[env]; !ok {
			output[env] = make(map[string]string)
		}

		if _, ok := output[env][app]; !ok {
			output[env][app] = ""
		}

		output[env][app] = output[env][app] + fmt.Sprintf("\t\tsecret \"%s\" {\n", secret.Key)

		for k, v := range secret.Secret.Data {
			switch vv := v.(type) {
			case string:
				output[env][app] = output[env][app] + fmt.Sprintf("\t\t\t%s = \"%s\"\n", k, escapeValue(vv))
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

			file := "conf.d/" + _env + "_" + _app + ".hcl"
			err := ioutil.WriteFile(file, []byte(content), 0644)
			if err != nil {
				return err
			}
			log.Infof("Wrote file: %s", file)

			output, err := formatHclFile(file)
			if err != nil {
				return err
			}

			log.Infof("Formatted file: %s (%s)", file, output.String())
		}
	}

	return nil
}

func escapeValue(value string) string {
	value = strings.Replace(value, "\\", "\\\\", -1)
	value = strings.Replace(value, "\"", "\\", -1)

	return value
}

func formatHclFile(file string) (bytes.Buffer, error) {
	cmd := exec.Command("hclfmt", "-w", file)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	return out, err
}
