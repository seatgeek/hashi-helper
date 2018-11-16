package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
	yaml "gopkg.in/yaml.v2"
)

type templater struct {
	templateVariables map[string]interface{}
}

func newTemplater(c *Config, variables, variableFiles []string) (*templater, error) {
	t := &templater{
		templateVariables: map[string]interface{}{},
	}

	if err := t.readTemplateVariablesFiles(variableFiles); err != nil {
		return nil, err
	}

	if err := t.parseTemplateVariables(variables); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *templater) renderContent(content, file string, depth int) (string, error) {
	log.Debugf("Rendering file %s (depth %d)", file, depth)

	if depth > 5 {
		return "", fmt.Errorf("recursive template rendering found, aborting")
	}

	fns := template.FuncMap{
		"service":                   t.service,
		"service_with_tag":          t.serviceWithTag,
		"grant_credentials":         t.grantCredentials,
		"grant_credentials_policy":  t.grantCredentialsPolicy,
		"github_assign_team_policy": t.githubAssignTeamPolicy,
		"ldap_assign_group_policy":  t.ldapAssignTeamPolicy,
		"lookup":                    t.lookupVar,
	}

	tmpl, err := template.New(file).
		Funcs(fns).
		Option("missingkey=error").
		Delims("[[", "]]").
		Parse(content)
	if err != nil {
		return "", err
	}

	// render the template to an internal buffer
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if err := tmpl.Execute(writer, t.templateVariables); err != nil {
		return "", err
	}

	// flush the buffer so we can read it out as a string
	if err := writer.Flush(); err != nil {
		return "", err
	}

	content = b.String()

	// check if we got any recursive rendering to do
	// we basically check if our delimiters exist in the file or not
	if strings.Contains(content, "[[") && strings.Contains(content, "]]") {
		return t.renderContent(content, file, depth+1)
	}

	fmt.Println("-----")
	fmt.Println(content)
	fmt.Println("-----")

	// HCL pretty print the rendered file
	res, err := printer.Format(b.Bytes())
	if err != nil {
		return "", fmt.Errorf("Could not format HCL file %s: %s", file, err)
	}

	// Trim the string for spaces / newlines and return the result
	return strings.TrimSpace(string(res)), nil
}

func (t *templater) parseTemplateVariables(pairs []string) error {
	for _, val := range pairs {
		chunks := strings.SplitN(val, "=", 2)
		if len(chunks) != 2 {
			return fmt.Errorf("Interpolation key/value pair '%s' is not valid", val)
		}

		t.templateVariables[chunks[0]] = chunks[1]
	}

	return nil
}

func (t *templater) readTemplateVariablesFiles(files []string) error {
	for _, variableFile := range files {
		ext := path.Ext(variableFile)

		var variables map[string]interface{}
		var err error

		switch ext {
		case ".hcl":
			variables, err = t.parseHCLVars(variableFile)
		case ".yaml", ".yml":
			variables, err = t.parseYAMLVars(variableFile)
		case ".json":
			variables, err = t.parseJSONVars(variableFile)
		default:
			err = fmt.Errorf("variables file extension %v not supported", ext)
		}

		if err != nil {
			return err
		}

		for k, v := range variables {
			t.templateVariables[k] = v
		}
	}

	return nil
}

func (t *templater) parseJSONVars(variableFile string) (variables map[string]interface{}, err error) {
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

func (t *templater) parseYAMLVars(variableFile string) (variables map[string]interface{}, err error) {
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

func (t *templater) parseHCLVars(variableFile string) (variables map[string]interface{}, err error) {
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
