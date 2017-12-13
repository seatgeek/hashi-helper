package vault

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	cli "gopkg.in/urfave/cli.v1"
	yaml "gopkg.in/yaml.v2"
)

type profiles map[string]profile
type profile struct {
	Token  string `yaml:"token"`
	Server string `yaml:"server"`
}

// UseProfile ...
func UseProfile(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("Please provide a profile name as first argument")
	}

	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("Missing profile name")
	}

	var profiles profiles
	dat, err := GetProfileConfig()
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(dat, &profiles); err != nil {
		return err
	}

	profile, ok := profiles[name]
	if !ok {
		return fmt.Errorf("No profile with the name '%s' was found", name)
	}

	fmt.Printf("export VAULT_ADDR=%s\n", profile.Server)
	fmt.Printf("export VAULT_TOKEN=%s\n", profile.Token)

	return nil
}

func GetProfileConfig() ([]byte, error) {
	cmd := exec.Command("keybase", "pgp", "decrypt", "--infile", getProfileFile())

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	fmt.Fprintf(os.Stderr, "\n# Starting keybase decrypt of %s\n\n", getProfileFile())
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to run keybase gpg decrypt: %s - %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

func getProfileFile() string {
	path := os.Getenv("HASHIHELPER_CONFIG_FILE")
	if path == "" {
		path = os.Getenv("HOME") + "/.hashi-helper-config.yml.pgp"
	}
	return path
}
