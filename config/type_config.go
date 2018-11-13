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
	Applications      Applications
	configDirectory   string
	ConsulKVs         ConsulKVs
	ConsulServices    ConsulServices
	Environments      Environments
	TargetEnvironment string
	templateVariables map[string]string
	VaultAuths        VaultAuths
	VaultMounts       VaultMounts
	VaultPolicies     VaultPolicies
	VaultSecrets      VaultSecrets
}

// NewConfig will create a new Config struct based on a directory
func NewConfig(path string) (*Config, error) {
	config := &Config{}

	if err := config.scanDirectory(path); err != nil {
		return nil, err
	}

	return config, nil
}

// NewConfigFromCLI will take a CLI context and create config from it
func NewConfigFromCLI(c *cli.Context) (*Config, error) {
	config := &Config{
		TargetEnvironment: c.GlobalString("environment"),
		configDirectory:   c.GlobalString("config-dir"),
	}

	if len(c.GlobalStringSlice("variables")) > 0 {
		if err := config.parseTemplateVariables(c.GlobalStringSlice("variables")); err != nil {
			return nil, err
		}
	}

	if c.GlobalString("config-file") != "" {
		return config, config.readAndProcess(c.GlobalString("config-file"))
	}

	return config, config.scanDirectory(c.GlobalString("config-dir"))
}

// scanDirectory ...
func (c *Config) scanDirectory(directory string) error {
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
			if err := c.readAndProcess(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, fmt.Errorf("[%s] %s", strings.TrimPrefix(directory, c.configDirectory)+"/"+fi.Name(), err))
			}

			continue
		}

		if fi.IsDir() {
			if err := c.scanDirectory(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, err)
			}

			continue
		}

		log.Debugf("Ignoring file %s/%s", directory, fi.Name())
	}

	return result
}

func (c *Config) readAndProcess(file string) error {
	content, err := c.readFile(file)
	if err != nil {
		return err
	}

	content, err = c.renderContent(content)
	if err != nil {
		return err
	}

	list, err := c.parseContent(content)
	if err != nil {
		return err
	}

	return c.processContent(list)
}

// Read File Content
func (c *Config) readFile(file string) (string, error) {
	log.Debugf("Parsing file %s", file)

	// read file from disk
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (c *Config) renderContent(content string) (string, error) {
	log.Debug("Rendering content")
	// create a template from the file content

	fns := template.FuncMap{
		"service":          c.service,
		"service_with_tag": c.serviceWithTag,
	}

	tmpl, err := template.New("<file>").
		Funcs(fns).
		Option("missingkey=error").
		Delims("[[", "]]").
		Parse(content)
	if err != nil {
		return "", err
	}

	// render the template
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if err := tmpl.Execute(writer, c.templateVariables); err != nil {
		return "", err
	}
	writer.Flush()

	return b.String(), nil
}

func (c *Config) parseContent(content string) (*ast.ObjectList, error) {
	// Parse into HCL AST
	log.Debug("Parsing content")
	root, hclErr := hcl.Parse(content)
	if hclErr != nil {
		return nil, fmt.Errorf("Could not parse content: %s", hclErr)
	}

	res, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return nil, fmt.Errorf("error parsing: root should be an object")
	}

	return res, nil
}

func (c *Config) processContent(list *ast.ObjectList) error {
	return c.processEnvironments(list)
}

func (c *Config) parseTemplateVariables(pairs []string) error {
	c.templateVariables = map[string]string{}

	for _, val := range pairs {
		chunks := strings.SplitN(val, "=", 2)
		if len(chunks) != 2 {
			return fmt.Errorf("Interpolation key/value pair '%s' is not valid", val)
		}

		c.templateVariables[chunks[0]] = chunks[1]
	}

	return nil
}
