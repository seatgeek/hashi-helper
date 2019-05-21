package profile

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	cli "gopkg.in/urfave/cli.v1"
	yaml "gopkg.in/yaml.v2"
)

type profiles map[string]profile

type authConfig struct {
	Method      string `yaml:"method"`
	CredsPath   string `yaml:"creds_path"`
	Token       string `yaml:"token"`
	UnsealToken string `yaml:"unseal_token"`
	GithubToken string `yaml:"github_token"`
}

type vaultCreds struct {
	Auth   authConfig `yaml:"auth"`
	Server string     `yaml:"server"`
}

type consulCreds struct {
	Server string     `yaml:"server"`
	Auth   authConfig `yaml:"auth"`
}

type nomadCreds struct {
	Auth   authConfig `yaml:"auth"`
	Server string     `yaml:"server"`
}
type profile struct {
	Vault  vaultCreds  `yaml:"vault"`
	Consul consulCreds `yaml:"consul"`
	Nomad  nomadCreds  `yaml:"nomad"`
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

	if profile.Vault.Server != "" {
		fmt.Printf("export VAULT_ADDR=%s\n", profile.Vault.Server)
	}

	if profile.Vault.Auth.Token != "" {
		fmt.Printf("export VAULT_TOKEN=%s\n", profile.Vault.Auth.Token)
	}

	if profile.Vault.Auth.UnsealToken != "" {
		fmt.Printf("export VAULT_UNSEAL_KEY=%s\n", profile.Vault.Auth.UnsealToken)
	}

	if profile.Vault.Auth.Method != "" {
		switch profile.Vault.Auth.Method {
		case "github":
			if profile.Vault.Auth.GithubToken == "" {
				return fmt.Errorf("github_token should be provided when using GitHub Vault auth method")
			} else {
				fmt.Printf("vault login -no-print -method=github token=%s\n", profile.Vault.Auth.GithubToken)
			}
			// More Auth methods to be added here
		}
	}

	if profile.Consul.Server != "" {
		fmt.Printf("export CONSUL_HTTP_ADDR=%s\n", profile.Consul.Server)
	}

	if profile.Consul.Auth.Token != "" {
		fmt.Printf("export CONSUL_HTTP_TOKEN=%s\n", profile.Consul.Auth.Token)
	}

	if profile.Consul.Auth.Method == "vault" {
		fmt.Printf("export CONSUL_HTTP_TOKEN=$(vault read -field=secret_id %s)\n", profile.Consul.Auth.CredsPath)

	}

	if profile.Nomad.Server != "" {
		fmt.Printf("export NOMAD_ADDR=%s\n", profile.Nomad.Server)
	}

	if profile.Nomad.Auth.Token != "" {
		fmt.Printf("export NOMAD_TOKEN=%s\n", profile.Nomad.Auth.Token)
	}

	if profile.Nomad.Auth.Method == "vault" {
		fmt.Printf("export NOMAD_TOKEN=$(vault read -field=secret_id %s)\n", profile.Nomad.Auth.CredsPath)

	}

	return nil
}

func getProfileConfig() ([]byte, error) {
	cmd := exec.Command("keybase", "pgp", "decrypt", "--infile", getProfileFile())

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	fmt.Fprintf(os.Stdout, "# Starting keybase decrypt of %s\n", getProfileFile())
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to run keybase gpg decrypt: %s - %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

func getProfileFile() string {
	path := os.Getenv("HASHI_HELPER_PROFILE_FILE")
	if path == "" {
		path = os.Getenv("HOME") + "/.hashi_helper_profiles.pgp"
	}
	return path
}
