package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"gopkg.in/urfave/cli.v1"
	yaml "gopkg.in/yaml.v2"
)

// Config ...
type Config struct {
	Applications      Applications
	configDirectory   string
	ConsulKVs         ConsulKVs
	ConsulServices    ConsulServices
	Environments      Environments
	TargetEnvironment string
	templateVariables map[string]interface{}
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

	if len(c.GlobalStringSlice("variable")) > 0 {
		if err := config.parseTemplateVariables(c.GlobalStringSlice("variable")); err != nil {
			return nil, err
		}
	}

	if len(c.GlobalStringSlice("variable-file")) > 0 {
		if err := config.readTemplateVariablesFiles(c.GlobalStringSlice("variable-file")); err != nil {
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
	if strings.HasSuffix(file, ".var.hcl") {
		log.Warnf("Ignoring files with .var.hcl extension")
		return nil
	}

	content, err := c.readFile(file)
	if err != nil {
		return err
	}

	content, err = c.renderContent(content, 0)
	if err != nil {
		return err
	}

	log.WithField("file", file).Debug(content)

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

func (c *Config) renderContent(content string, depth int) (string, error) {
	log.Debugf("Rendering content depth %d", depth)

	if depth > 5 {
		return "", fmt.Errorf("recursive template rendering found, aborting")
	}

	fns := template.FuncMap{
		"service":                   c.service,
		"service_with_tag":          c.serviceWithTag,
		"grant_credentials":         c.grantCredentials,
		"grant_credentials_policy":  c.grantCredentialsPolicy,
		"github_assign_team_policy": c.githubAssignTeamPolicy,
		"ldap_assign_group_policy":  c.ldapAssignTeamPolicy,
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

	if strings.Contains(b.String(), "[[") {
		return c.renderContent(b.String(), depth+1)
	}

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
	c.templateVariables = map[string]interface{}{}

	for _, val := range pairs {
		chunks := strings.SplitN(val, "=", 2)
		if len(chunks) != 2 {
			return fmt.Errorf("Interpolation key/value pair '%s' is not valid", val)
		}

		c.templateVariables[chunks[0]] = chunks[1]
	}

	return nil
}

func (c *Config) readTemplateVariablesFiles(files []string) error {
	for _, variableFile := range files {
		ext := path.Ext(variableFile)

		var variables map[string]interface{}
		var err error

		switch ext {
		case ".hcl":
			variables, err = c.parseHCLVars(variableFile)
		case ".yaml", ".yml":
			variables, err = c.parseYAMLVars(variableFile)
		case ".json":
			variables, err = c.parseJSONVars(variableFile)
		default:
			err = fmt.Errorf("variables file extension %v not supported", ext)
		}

		if err != nil {
			return err
		}

		for k, v := range variables {
			c.templateVariables[k] = v
		}
	}

	return nil
}

func (c *Config) parseJSONVars(variableFile string) (variables map[string]interface{}, err error) {
	jsonFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return
	}

	variables = make(map[string]interface{})
	if err = json.Unmarshal(jsonFile, &variables); err != nil {
		return
	}

	return variables, nil
}

func (c *Config) parseYAMLVars(variableFile string) (variables map[string]interface{}, err error) {
	yamlFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return
	}

	variables = make(map[string]interface{})
	if err = yaml.Unmarshal(yamlFile, &variables); err != nil {
		return
	}

	return variables, nil
}

func (c *Config) parseHCLVars(variableFile string) (variables map[string]interface{}, err error) {
	hclFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return
	}

	variables = make(map[string]interface{})
	if err := hcl.Decode(&variables, string(hclFile)); err != nil {
		return nil, err
	}

	return variables, nil
}
