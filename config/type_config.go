package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"gopkg.in/urfave/cli.v1"
)

// Config ...
type Config struct {
	TargetEnvironment string
	ConfigDirectory   string
	Interpolations    map[string]string
	Applications      Applications
	Environments      Environments
	VaultMounts       VaultMounts
	VaultPolicies     VaultPolicies
	VaultSecrets      VaultSecrets
	VaultAuths        VaultAuths
	ConsulServices    ConsulServices
	ConsulKVs         ConsulKVs
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
	config := &Config{
		TargetEnvironment: c.GlobalString("environment"),
		ConfigDirectory:   c.GlobalString("config-dir"),
	}

	if len(c.GlobalStringSlice("interpolation")) > 0 {
		if err := config.buildInterpolation(c.GlobalStringSlice("interpolation")); err != nil {
			return nil, err
		}
	}

	if c.GlobalString("config-file") != "" {
		return config, config.ReadAndProcess(c.GlobalString("config-file"))
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
			if err := c.ReadAndProcess(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, fmt.Errorf("[%s] %s", strings.TrimPrefix(directory, c.ConfigDirectory)+"/"+fi.Name(), err))
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

func (c *Config) ReadAndProcess(file string) error {
	content, err := c.ReadFile(file)
	if err != nil {
		return err
	}

	list, err := c.ParseContent(content)
	if err != nil {
		return err
	}

	return c.ProcessContent(list)
}

// Read File Content
func (c *Config) ReadFile(file string) (string, error) {
	log.Debugf("Parsing file %s", file)

	// read file from disk
	configContent, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	// create a template from the file content
	tmpl, err := template.New(file).
		Option("missingkey=error").
		Parse(string(configContent))
	if err != nil {
		return "", err
	}

	// render the template
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if err := tmpl.Execute(writer, c.Interpolations); err != nil {
		return "", err
	}
	writer.Flush()

	// return the template string
	return b.String(), nil
}

func (c *Config) ParseContent(configContent string) (*ast.ObjectList, error) {
	log.Debug("Parsing content")

	// Parse into HCL AST
	root, hclErr := hcl.Parse(configContent)
	if hclErr != nil {
		return nil, fmt.Errorf("Could not parse content: %s", hclErr)
	}

	res, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return nil, fmt.Errorf("error parsing: root should be an object")
	}

	return res, nil
}

func (c *Config) ProcessContent(list *ast.ObjectList) error {
	return c.processEnvironments(list)
}

func (c *Config) buildInterpolation(pairs []string) error {
	c.Interpolations = map[string]string{}

	for _, val := range pairs {
		chunks := strings.SplitN(val, "=", 2)
		if len(chunks) != 2 {
			return fmt.Errorf("Interpolation key/value pair '%s' is not valid", val)
		}

		c.Interpolations[chunks[0]] = chunks[1]
	}

	return nil
}
