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
	Token        string `yaml:"token"`
	UnsealToken  string `yaml:"unseal_token"`
	Server       string `yaml:"server"`
	ConsulServer string `yaml:"consul_server"`
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
	dat, err := getProfileConfig()
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

	if profile.Server != "" {
		fmt.Printf("export VAULT_ADDR=%s\n", profile.Server)
	}

	if profile.ConsulServer != "" {
		fmt.Printf("export CONSUL_HTTP_ADDR=%s\n", profile.ConsulServer)
	}

	if profile.Token != "" {
		fmt.Printf("export VAULT_TOKEN=%s\n", profile.Token)
	}

	if profile.UnsealToken != "" {
		fmt.Printf("export VAULT_UNSEAL_KEY=%s\n", profile.UnsealToken)
	}

	return nil
}

func getProfileConfig() ([]byte, error) {
	cmd := exec.Command("keybase", "pgp", "decrypt", "--infile", getProfileFile())

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	fmt.Fprintf(os.Stderr, "# Starting keybase decrypt of %s\n", getProfileFile())
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to run keybase gpg decrypt: %s - %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

func getProfileFile() string {
	path := os.Getenv("VAULT_PROFILE_FILE")
	if path == "" {
		path = os.Getenv("HOME") + "/.vault_profiles.pgp"
	}
	return path
}
