package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"strings"

	log "github.com/Sirupsen/logrus"
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
	ConsulServices ConsulServices
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
	defer d.Close()

	fi, err := d.Readdir(-1)
	if err != nil {
		return err
	}

	for _, fi := range fi {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".hcl") {
			if err := c.AddFile(directory + "/" + fi.Name()); err != nil {
				return err
			}

			continue
		}

		if fi.IsDir() {
			if err := c.ScanDirectory(directory + "/" + fi.Name()); err != nil {
				return err
			}

			continue
		}

		log.Debugf("Ignoring file %s/%s", directory, fi.Name())
	}

	return nil
}

// AddFile to the config struct
func (c *Config) AddFile(file string) error {
	log.Debugf("Parsing file %s", file)

	configContent, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	// Parse into HCL AST
	root, hclErr := hcl.Parse(string(configContent))
	if hclErr != nil {
		return hclErr
	}

	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return fmt.Errorf("error parsing: root should be an object")
	}

	return c.processEnvironments(list)
}
