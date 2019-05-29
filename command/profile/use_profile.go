package profile

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/vault/api"
	vault "github.com/hashicorp/vault/api"
	vgh "github.com/hashicorp/vault/builtin/credential/github"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

type profiles map[string]profile

type InternalTokenHelper struct {
	tokenPath   string
	profileName string
}

type authConfig struct {
	Method      string `yaml:"method"`
	CredsPath   string `yaml:"creds_path"`
	Token       string `yaml:"token"`
	UnsealToken string `yaml:"unseal_token"`
	GithubToken string `yaml:"github_token"`
	GithubMount string `yaml:"mount"`
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

	// if vault_token file for this profile exists checking the saved creds

	i := InternalTokenHelper{}
	i.profileName = name
	v, err := api.NewClient(vault.DefaultConfig())
	if err != nil {
		return err
	}
	storedToken, err := i.Get()
	if err != nil {
		return err
	}

	if profile.Vault.Server != "" {
		v.SetAddress(profile.Vault.Server)
		fmt.Printf("export VAULT_ADDR=%s\n", profile.Vault.Server)
	}

	if profile.Vault.Auth.Token != "" {
		if storedToken != profile.Vault.Auth.Token {
			fmt.Printf("Token stored in %s is different from the Token stored in %s for profile %s", i.tokenPath, getProfileFile(), name)
			secret, _ := v.Auth().Token().LookupSelf()
			if ttl, _ := secret.TokenTTL(); ttl > 0 {
				v.Auth().Token().RenewSelf(0)
			}
		} else {

		}
		v.SetToken(storedToken)

		fmt.Printf("export VAULT_TOKEN=%s\n", profile.Vault.Auth.Token)
		// TODO: check token is valid
	}

	if profile.Vault.Auth.UnsealToken != "" {
		fmt.Printf("export VAULT_UNSEAL_KEY=%s\n", profile.Vault.Auth.UnsealToken)
	}

	if profile.Vault.Auth.Method != "" {
		switch profile.Vault.Auth.Method {
		case "github":
			// TODO: check stored token - if it exists is valid and issued using github
			m := make(map[string]string)
			if profile.Vault.Auth.GithubMount != "" {
				m["mount"] = profile.Vault.Auth.GithubMount
			}
			if profile.Vault.Auth.GithubToken == "" {
				return fmt.Errorf("github_token should be provided when using GitHub Vault auth method")
			} else {
				m["token"] = profile.Vault.Auth.GithubToken
				h := vgh.CLIHandler{}
				secret, err := h.Auth(v, m)
				if err != nil {
					return err
				}

				os.Setenv("VAULT_TOKEN", secret.Auth.ClientToken)
				fmt.Printf("export VAULT_AUTH_GITHUB_TOKEN=%s\n", profile.Vault.Auth.GithubToken)

				fmt.Printf("export VAULT_TOKEN=$(vault login -field token -method=github)\n", profile.Vault.Auth.GithubToken)
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
		homePath, err := homedir.Dir()
		if err != nil {
			panic(fmt.Sprintf("error getting user's home directory: %v", err))
		}
		path = filepath.Join(homePath, "/.hashi_helper_profiles.pgp")
	}
	return path
}

func getVaultTokenFile(profileName string) string {
	homePath, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("error getting user's home directory: %v", err))
	}
	path := filepath.Join(homePath, "/.vault_token_", profileName)

	return path
}

// populateTokenPath figures out the token path using homedir to get the user's
// home directory
func (i *InternalTokenHelper) populateTokenPath() {
	homePath, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("error getting user's home directory: %v", err))
	}
	i.tokenPath = filepath.Join(homePath, ".vault-token-", i.profileName)
}

func (i *InternalTokenHelper) Path() string {
	return i.tokenPath
}

// Get gets the value of the stored token, if any
func (i *InternalTokenHelper) Get() (string, error) {
	i.populateTokenPath()
	f, err := os.Open(i.tokenPath)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, f); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// Store stores the value of the token to the file
func (i *InternalTokenHelper) Store(input string) error {
	i.populateTokenPath()
	f, err := os.OpenFile(i.tokenPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bytes.NewBufferString(input)
	if _, err := io.Copy(f, buf); err != nil {
		return err
	}

	return nil
}

// Erase erases the value of the token
func (i *InternalTokenHelper) Erase() error {
	i.populateTokenPath()
	if err := os.Remove(i.tokenPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
