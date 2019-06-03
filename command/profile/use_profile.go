package profile

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hashicorp/vault/api"
	vault "github.com/hashicorp/vault/api"
	vgh "github.com/hashicorp/vault/builtin/credential/github"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

type profiles map[string]profileStruct

type profileStruct struct {
	Vault  vaultCreds  `yaml:"vault"`
	Consul consulCreds `yaml:"consul"`
	Nomad  nomadCreds  `yaml:"nomad"`
}

type InternalTokenHelper struct {
	tokenPath   string
	profileName string
}

type authConfig struct {
	Method      string `yaml:"method"`
	CredsPath   string `yaml:"creds_path"`
	Token       string `yaml:"token"`
	ExpireTime  string `yaml:"expire_time"` //used internal for cache files only
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

// UseProfile ...
func UseProfile(c *cli.Context) error {

	if !c.Args().Present() {
		return fmt.Errorf("Please provide a profile name as first argument")
	}

	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("Missing profile name")
	}

	// parsing profiles file
	var parsedProfiles, profilesCache profiles
	dat, err := decryptFile(getProfileFile())
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(dat, &parsedProfiles); err != nil {
		return err
	}

	profile, ok := parsedProfiles[name]
	if !ok {
		return fmt.Errorf("No profile with the name '%s' was found", name)
	}

	cacheDat, err := decryptFile(getCacheFile())

	if cacheDat != nil {
		if err := yaml.Unmarshal(cacheDat, &profilesCache); err != nil {
			return err
		}
	} else {
		profilesCache = make(profiles)
	}

	profileCache := profilesCache[name]

	// Creating Vault Client for checking creds
	v, err := api.NewClient(vault.DefaultConfig())
	// setting Vault timeout to 5 seconds
	v.SetClientTimeout(time.Second * 5)

	if profile.Vault.Server != "" {
		v.SetAddress(profile.Vault.Server)
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
			if profileCache.Vault.Auth.Token != "" {
				v.SetToken(profileCache.Vault.Auth.Token)
				if profileCache.Vault.Auth.ExpireTime != "" {
					et, err := time.Parse(time.RFC3339Nano, profileCache.Vault.Auth.ExpireTime)
					if err != nil {
						fmt.Errorf("Bad vault:auth:expire_time in cache for %s profile ", name)
					}
					if et.Before(time.Now()) {
						profileCache.Vault, err = vaultLoginGitHub(profile.Vault, v)
						if err != nil {
							fmt.Printf("Bad github login")
						}
					}
				}
			} else {
				profileCache.Vault, err = vaultLoginGitHub(profile.Vault, v)
				if err != nil {
					fmt.Printf("Bad github login")
				}
			}
			v.SetToken(profileCache.Vault.Auth.Token)
			fmt.Printf("export VAULT_TOKEN=%s\n", profileCache.Vault.Auth.Token)
			// TODO: More Auth methods to be added here
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
		if profileCache.Nomad.Auth.Token != "" {
			if profileCache.Nomad.Auth.ExpireTime != "" {
				et, err := time.Parse(time.RFC3339Nano, profileCache.Nomad.Auth.ExpireTime)
				if err != nil {
					fmt.Errorf("Bad nomad:auth:expire_time in cache for %s profile ", name)
				}
				if et.Before(time.Now()) {
					profileCache.Nomad, err = vaultGetNomadCreds(profile.Nomad, v)
				}
			}
		} else {
			profileCache.Nomad, err = vaultGetNomadCreds(profile.Nomad, v)
		}
		fmt.Printf("export NOMAD_TOKEN=%s\n", profileCache.Nomad.Auth.Token)
	}

	profilesCache[name] = profileCache

	// Create a file for IO
	cacheTemp, err := ioutil.TempFile("", "hashi_helper_cache")
	if err != nil {
		panic(err)
	}

	y, err := yaml.Marshal(&profilesCache)
	if err != nil {
		fmt.Errorf("bugga")
	}

	// Write to the file
	if err := ioutil.WriteFile(cacheTemp.Name(), y, 600); err != nil {
		panic(err)
	}
	cacheTemp.Close()

	defer os.Remove(cacheTemp.Name())

	encryptFile(cacheTemp.Name(), getCacheFile())

	return nil
}

func vaultGetNomadCreds(n nomadCreds, vc *vault.Client) (nomadCreds, error) {
	r, err := readFromVault(vc, n.Auth.CredsPath)
	if err != nil {
		return n, err
	}

	n.Auth.Token = r.Data["secret_id"].(string)
	n.Auth.ExpireTime = time.Now().Add(time.Second * time.Duration(r.LeaseDuration)).Format(time.RFC3339Nano)

	return n, err
}

func vaultLoginGitHub(v vaultCreds, vc *vault.Client) (vaultCreds, error) {
	m := make(map[string]string)
	if v.Auth.GithubMount != "" {
		m["mount"] = v.Auth.GithubMount
	}
	if v.Auth.GithubToken == "" {
		return v, fmt.Errorf("github_token should be provided when using GitHub Vault auth method")
	} else {
		m["token"] = v.Auth.GithubToken
		h := vgh.CLIHandler{}
		secret, err := h.Auth(vc, m)
		if err != nil {
			return v, err
		}

		v.Auth.Token = secret.Auth.ClientToken
		v.Auth.ExpireTime = time.Now().Add(time.Second * time.Duration(secret.Auth.LeaseDuration)).Format(time.RFC3339Nano)

		return v, nil
	}
}

func readFromVault(v *vault.Client, path string) (*vault.Secret, error) {

	creds, err := v.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

func getCacheFile() string {
	path := os.Getenv("HASHI_HELPER_CACHE_FILE")
	if path == "" {
		homePath, err := homedir.Dir()
		if err != nil {
			panic(fmt.Sprintf("error getting user's home directory: %v", err))
		}
		path = filepath.Join(homePath, "/.hashi_helper_cache.pgp")
	}
	return path
}

func decryptFile(filePath string) ([]byte, error) {

	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}
	cmd := exec.Command("keybase", "pgp", "decrypt", "--infile", filePath)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	fmt.Fprintf(os.Stdout, "# Starting keybase decrypt of %s\n", filePath)
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to run keybase gpg decrypt: %s - %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

func encryptFile(inFile, outFile string) error {

	encryptCmd := exec.Command("keybase", "pgp", "encrypt", "--infile", inFile, "--outfile", outFile)

	encryptErr := encryptCmd.Run()
	if encryptErr != nil {
		return encryptErr
	}

	return nil

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
