package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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

func (t *templater) consulDomain() (string, error) {
	val, ok := t.templateVariables["consul_domain"]
	if !ok {
		return "", errors.New("Missing interpolation key 'consul_domain'")
	}

	return fmt.Sprintf("%s", val), nil
}

func (t *templater) lookupVar(key string) (interface{}, error) {
	val, ok := t.templateVariables[key]
	if !ok {
		return "", fmt.Errorf("Missing interpolation key '%s'", key)
	}
	return val, nil
}

func (t *templater) service(service string) (interface{}, error) {
	domain, err := t.consulDomain()
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("%s.service.%s", service, domain), nil
}

func (t *templater) serviceWithTag(service, tag string) (interface{}, error) {
	domain, err := t.consulDomain()
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("%s.%s.service.%s", tag, service, domain), nil
}

func (t *templater) grantCredentials(db, role string) (interface{}, error) {
	tmpl := `
path "%s/creds/%s" {
  capabilities = ["read"]
}`

	return fmt.Sprintf(tmpl, db, role), nil
}

func (t *templater) grantCredentialsPolicy(db, role string) (interface{}, error) {
	tmpl := `
policy "%s-%s" {
	[[ grant_credentials "%s" "%s" ]]
}`

	return fmt.Sprintf(tmpl, db, role, db, role), nil
}

func (t *templater) githubAssignTeamPolicy(team, policy string) (interface{}, error) {
	tmpl := `
secret "/auth/github/map/teams/%s" {
  value = "%s"
}`

	return fmt.Sprintf(tmpl, team, policy), nil
}

func (t *templater) ldapAssignTeamPolicy(group, policy string) (interface{}, error) {
	tmpl := `
secret "/auth/ldap/groups/%s" {
  value = "%s"
}`

	return fmt.Sprintf(tmpl, group, policy), nil
}