package config

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

// DefaultConcurrency ...
var DefaultConcurrency int

// TargetEnvironment ...
var TargetEnvironment string

// TargetApplication ...
var TargetApplication string

// Config ...
type Config struct {
	templater         *templater
	logger            *log.Entry
	Applications      Applications
	configDirectory   string
	ConsulKVs         ConsulKVs
	ConsulServices    ConsulServices
	Environments      Environments
	TargetEnvironment string
	VaultAuths        VaultAuths
	VaultMounts       VaultMounts
	VaultPolicies     VaultPolicies
	VaultSecrets      VaultSecrets
}

// NewConfigFromCLI will take a CLI context and create config from it
func NewConfigFromCLI(c *cli.Context) (*Config, error) {
	config := &Config{
		TargetEnvironment: c.GlobalString("environment"),
	}

	// create a templater we can use for future rendering
	templater, err := newTemplater(c.GlobalStringSlice("variable"), c.GlobalStringSlice("variable-file"))
	if err != nil {
		return nil, err
	}

	// scan all config-dirs provided
	for _, dir := range c.GlobalStringSlice("config-dir") {
		scanner := newConfigScanner(dir, config, templater)
		if err := scanner.scan(); err != nil {
			return nil, err
		}
	}

	// scan all config-files provided
	for _, file := range c.GlobalStringSlice("config-file") {
		scanner := newConfigScanner(file, config, templater)
		if err := scanner.scan(); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func (c *Config) parseContent(content, file string) (*ast.ObjectList, error) {
	// Parse into HCL AST
	log.WithField("file", file).Debug("Parsing content")
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

func (c *Config) processContent(list *ast.ObjectList, file string) error {
	c.logger = log.WithField("file", file)
	defer func() {
		c.logger = log.WithField("file", "")
	}()

	return c.processEnvironments(list)
}

// c.checkHCLKeys
// Simply checks if there is any unexpected keys in the AST node provided, nice way to avoid a typo
func (c *Config) checkHCLKeys(node ast.Node, valid []string) error {
	var list *ast.ObjectList
	switch n := node.(type) {
	case *ast.ObjectList:
		list = n
	case *ast.ObjectType:
		list = n.List
	default:
		return fmt.Errorf("cannot check HCL keys of type %T", n)
	}

	validMap := make(map[string]struct{}, len(valid))
	for _, v := range valid {
		validMap[v] = struct{}{}
	}

	var result error
	for _, item := range list.Items {
		key := item.Keys[0].Token.Value().(string)
		if _, ok := validMap[key]; !ok {
			result = multierror.Append(result, fmt.Errorf(
				"invalid key '%s' in line %+v", key, item.Keys[0].Token.Pos))
		}
	}

	return result
}
