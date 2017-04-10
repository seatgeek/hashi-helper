package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/mapstructure"
)

var DefaultConcurrency int

// NewConfigFile will return a Config struct
func NewConfigFile(file string) (Environments, error) {
	// Read file
	configContent, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Parse into HCL AST
	root, hclErr := hcl.Parse(string(configContent))
	if hclErr != nil {
		return nil, hclErr
	}

	// Top-level item should be a list
	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return nil, fmt.Errorf("error parsing: root should be an object")
	}

	return parseEnvironmentStanza(list)
}

// NewConfigFromDirectory ...
func NewConfigFromDirectory(dirname string) (Environments, error) {
	d, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	fi, err := d.Readdir(-1)
	if err != nil {
		return nil, err
	}

	root := make(Environments)

	for _, fi := range fi {
		if fi.Mode().IsRegular() {
			c, e := NewConfigFile(dirname + "/" + fi.Name())
			if e != nil {
				return nil, e
			}

			root.merge(c)
		}
	}

	return root, nil
}

// Parse root environment stanza
func parseEnvironmentStanza(list *ast.ObjectList) (Environments, error) {
	// Check for invalid or unknown root keys
	valid := []string{"environment"}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, err
	}

	config := make(Environments)
	for _, env := range list.Items {
		if len(env.Keys) != 2 {
			return nil, fmt.Errorf("Missing environment name in line %+v", env.Keys[0].Pos())
		}

		envName := env.Keys[1].Token.Value().(string)

		res, err := parseApplicationStanza(env.Val.(*ast.ObjectType).List)
		if err != nil {
			return nil, err
		}

		if _, ok := config[envName]; !ok {
			config[envName] = res
		} else {
			config[envName].merge(res)
		}
	}

	return config, nil
}

// parseEnvironmentStanza
// parse out `environment -> application {)` stanza
func parseApplicationStanza(list *ast.ObjectList) (Environment, error) {
	valid := []string{"application"}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, err
	}

	matches := list.Filter("application")
	if len(matches.Items) == 0 {
		return nil, fmt.Errorf("no 'application' stanza found")
	}

	env := make(Environment)
	for _, app := range list.Items {
		if len(app.Keys) != 2 {
			return nil, fmt.Errorf("Missing application name in line %+v", app.Keys[0].Pos())
		}

		appName := app.Keys[1].Token.Value().(string)

		res, err := parseSecretStanza(app.Val.(*ast.ObjectType).List)
		if err != nil {
			return nil, err
		}

		if _, ok := env[appName]; !ok {
			z := env[appName]
			z.Secrets = res
			env[appName] = z
		} else {
			env[appName].Secrets.merge(res)
		}
	}

	return env, nil
}

// parseSecretStanza
// parse out `environment -> application -> secret {}` stanza
func parseSecretStanza(list *ast.ObjectList) (Secrets, error) {
	// Check for invalid or unknown root keys
	valid := []string{"secret"}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, err
	}

	// Find all nomad stanzas
	matches := list.Filter("secret")
	if len(matches.Items) == 0 {
		return nil, fmt.Errorf("no 'secret' stanza found")
	}

	secrets := make(Secrets)

	for _, secret := range list.Items {
		if len(secret.Keys) != 2 {
			return nil, fmt.Errorf("Missing secret name in line %+v", secret.Keys[0].Pos())
		}

		secretName := secret.Keys[1].Token.Value().(string)

		var m map[string]interface{}
		if err := hcl.DecodeObject(&m, secret.Val); err != nil {
			return nil, err
		}

		var secret Secret
		if err := mapstructure.WeakDecode(m, &secret); err != nil {
			return nil, err
		}

		secrets[secretName] = secret
	}

	return secrets, nil
}

// checkHCLKeys
// Simply checks if there is any unexpected keys in the AST node provided, nice way to avoid a typo
func checkHCLKeys(node ast.Node, valid []string) error {
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
