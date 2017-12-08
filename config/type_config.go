package config

import (
  "bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"strings"

	log "github.com/Sirupsen/logrus"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"gopkg.in/urfave/cli.v1"
)

// Config ...
type Config struct {
	Applications   Applications
	Environments   Environments
	VaultMounts    VaultMounts
	VaultPolicies  VaultPolicies
	VaultSecrets   VaultSecrets
	VaultAuths     VaultAuths
	ConsulServices ConsulServices
	ConsulKVs      ConsulKVs
}

// NewConfig will create a new Config struct based on a directory
func NewConfig(path string) (*Config, error) {
	config := &Config{}

	if err := config.ScanDirectory(path); err != nil {
		return nil, err
	}

	return config, nil
}

// NewConfigFromCLI will take a CLI context and create config from it
func NewConfigFromCLI(c *cli.Context) (*Config, error) {
	config := &Config{}

	if c.GlobalString("config-file") != "" {
		return config, config.AddFile(c.GlobalString("config-file"))
	}

	return config, config.ScanDirectory(c.GlobalString("config-dir"))
}

// ScanDirectory ...
func (c *Config) ScanDirectory(directory string) error {
	log.Debugf("Scanning directory %s", directory)

	d, err := os.Open(directory)
	if err != nil {
		return err
	}
	d.Close()

	fi, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	var result error
	for _, fi := range fi {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".hcl") {
			if err := c.AddFile(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, fmt.Errorf("[%s] %s", directory+"/"+fi.Name(), err))
			}

			continue
		}

		if fi.IsDir() {
			if err := c.ScanDirectory(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, err)
			}

			continue
		}

		log.Debugf("Ignoring file %s/%s", directory, fi.Name())
	}

	return result
}

// AddFile to the config struct
func (c *Config) AddFile(file string) error {

  var encryptedHcl bool = false
	var configContent string = ""
	configContentRaw, err := ioutil.ReadFile(file)

	// look for file contents indicating an encrypted payload
	if strings.Contains(string(configContentRaw), "-----BEGIN PGP MESSAGE-----") {
		encryptedHcl = true
	}

	if encryptedHcl == true {
		// files encrypted with Keybase pgp (similar to Vault profiles file)
		log.Debugf("Decrypting file %s", file)
		cmd := exec.Command("keybase", "pgp", "decrypt", "--infile", file)

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Failed to run keybase gpg decrypt: %s - %s", err, stderr.String())
		}

		configContent = string(stdout.Bytes())

	} else {
		// unencrypted files
		log.Debugf("Parsing file %s", file)
		configContent = string(configContentRaw)
		if err != nil {
			return err
		}
	}

	// Parse into HCL AST
	root, hclErr := hcl.Parse(configContent)
	if hclErr != nil {
		return fmt.Errorf("Could not parse file %s: %s", file, hclErr)
	}

	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return fmt.Errorf("error parsing: root should be an object")
	}

	return c.processEnvironments(list)
}
