package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// DefaultConcurrency ...
var DefaultConcurrency int

// TargetEnvironment ...
var TargetEnvironment string

// TargetApplication ...
var TargetApplication string

// NewConfigFile will return a Config struct
func NewConfigFile(file string) (*Environments, error) {
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

	return processEnvironments(list)
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

			root.merge(*c)
		}

		if fi.IsDir() {
			sub, e := NewConfigFromDirectory(dirname + "/" + fi.Name())
			if err != nil {
				return nil, e
			}

			root.merge(sub)
		}
	}

	return root, nil
}
