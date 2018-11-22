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

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type renderer struct {
	variables       map[string]interface{}
	templateScratch *Scratch
	scratch         *Scratch
}

func newRenderer(variables, variableFiles []string) (*renderer, error) {
	t := &renderer{
		variables: map[string]interface{}{},
		scratch:   &Scratch{},
	}

	if err := t.readTemplateVariablesFiles(variableFiles); err != nil {
		return nil, err
	}

	if err := t.parseTemplateVariables(variables); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *renderer) renderContent(content, file string, depth int) (string, error) {
	log.Debugf("Rendering file %s (depth %d)", file, depth)

	if depth > 10 {
		return "", fmt.Errorf("recursive template rendering found, aborting")
	}

	fns := template.FuncMap{
		"consul_domain":             t.consulDomain,
		"github_assign_team_policy": t.githubAssignTeamPolicy,
		"grant_credentials_policy":  t.grantCredentialsPolicy,
		"grant_credentials":         t.grantCredentials,
		"ldap_assign_group_policy":  t.ldapAssignTeamPolicy,
		"lookup_default":            t.lookupVarDefault,
		"lookup_map_default":        t.lookupVarMapDefault,
		"lookup_map":                t.lookupVarMap,
		"lookup":                    t.lookupVar,
		"replace_all":               t.replaceAll,
		"scratch":                   t.createScratch(),
		"service_with_tag":          t.consulServiceWithTag,
		"service":                   t.consulService,
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
	if err := tmpl.Execute(writer, t.variables); err != nil {
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

func (t *renderer) parseTemplateVariables(pairs []string) error {
	for _, val := range pairs {
		chunks := strings.SplitN(val, "=", 2)
		if len(chunks) != 2 {
			return fmt.Errorf("Interpolation key/value pair '%s' is not valid", val)
		}

		t.variables[chunks[0]] = chunks[1]
	}

	return nil
}

func (t *renderer) readTemplateVariablesFiles(files []string) error {
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
			t.variables[k] = v
		}
	}

	return nil
}

// parseJSONVars will read a file from disk and JSON unmarshal it into a map[string]interface{}
func (t *renderer) parseJSONVars(variableFile string) (variables map[string]interface{}, err error) {
	jsonFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return nil, err
	}

	variables = make(map[string]interface{})
	if err = json.Unmarshal(jsonFile, &variables); err != nil {
		return nil, err
	}

	return variables, nil
}

// parseYAMLVars will read a file from disk and yaml unmarshal it into a map[string]interface{}
func (t *renderer) parseYAMLVars(variableFile string) (variables map[string]interface{}, err error) {
	yamlFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return nil, err
	}

	variables = make(map[string]interface{})
	if err = yaml.Unmarshal(yamlFile, &variables); err != nil {
		return nil, err
	}

	return variables, nil
}

// parseHCLVars will read a file from disk and hcl unmarshal it into a map[string]interface{}
func (t *renderer) parseHCLVars(variableFile string) (variables map[string]interface{}, err error) {
	hclFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return nil, err
	}

	variables = make(map[string]interface{})
	if err := hcl.Decode(&variables, string(hclFile)); err != nil {
		return nil, err
	}

	return variables, nil
}

// lookupVar will return the template variable identified by `key` or return an error
// which will abort the template rendering.
func (t *renderer) lookupVar(key string) (interface{}, error) {
	val, ok := t.variables[key]
	if !ok {
		return "", fmt.Errorf("Missing template variable '%s'", key)
	}

	return val, nil
}

// lookupVarDefault will return the template variable identified by `key` or a default value
// provided in `def`.
func (t *renderer) lookupVarDefault(key string, def interface{}) (interface{}, error) {
	val, ok := t.variables[key]
	if !ok {
		return def, nil
	}

	return val, nil
}

// lookupVarMap will return the value of "mapKey" within the template variable
// identified by "key"`.
//
// If "key" is not a template variable, an error will be returned
// If "key" is not a map[string]interface{}, an error will be returned
// if "mapKey" do not exist in the map of "key", an error will be returnedd
func (t *renderer) lookupVarMap(key, mapKey string) (interface{}, error) {
	if t.templateScratch == nil {
		t.templateScratch = &Scratch{values: t.variables}
	}

	return t.templateScratch.MapGet(key, mapKey)
}

// lookupVarMapDefault will return the value of "mapKey" within the template variable
// identified by "key"`.
//
// If "key" is not a template variable, an error will be returned
// If "key" is not a map[string]interface{}, an error will be returned
// if "mapKey" do not exist in the map of "key", the default value provided in "def" is returned.
func (t *renderer) lookupVarMapDefault(key, mapKey string, def interface{}) (interface{}, error) {
	v, err := t.lookupVarMap(key, mapKey)
	if err != nil {
		return def, nil
	}

	return v, nil
}

// consulDomain will return the Consul DNS Domain.
// It will default to "consul" unless template variable key "consul_domain" is defined
func (t *renderer) consulDomain() (interface{}, error) {
	return t.lookupVarDefault("consul_domain", "consul")
}

// consulService will return a Consul Service hostname
func (t *renderer) consulService(service string) (interface{}, error) {
	return fmt.Sprintf(`%s.service.[[ consul_domain ]]`, service), nil
}

// consulService will return a Consul Service with provided tag
func (t *renderer) consulServiceWithTag(service, tag string) (interface{}, error) {
	return fmt.Sprintf(`%s.%s.service.[[ consul_domain ]]`, tag, service), nil
}

func (t *renderer) grantCredentials(db, role string) (interface{}, error) {
	tmpl := `
path "%s/creds/%s" {
  capabilities = ["read"]
}`

	return fmt.Sprintf(tmpl, db, role), nil
}

func (t *renderer) grantCredentialsPolicy(db, role string) (interface{}, error) {
	tmpl := `
policy "%s-%s" {
	[[ grant_credentials "%s" "%s" ]]
}`

	return fmt.Sprintf(tmpl, db, role, db, role), nil
}

func (t *renderer) githubAssignTeamPolicy(team, policy string) (interface{}, error) {
	tmpl := `
secret "/auth/github/map/teams/%s" {
  value = "%s"
}`

	return fmt.Sprintf(tmpl, team, policy), nil
}

func (t *renderer) ldapAssignTeamPolicy(group, policy string) (interface{}, error) {
	tmpl := `
secret "/auth/ldap/groups/%s" {
  value = "%s"
}`

	return fmt.Sprintf(tmpl, group, policy), nil
}

func (t *renderer) createScratch() func() *Scratch {
	return func() *Scratch {
		if t.scratch == nil {
			t.scratch = &Scratch{}
		}
		return t.scratch
	}
}

// replaceAll replaces all occurrences of a value in a string with the given
// replacement value.
func (t *renderer) replaceAll(f, x, s string) (string, error) {
	return strings.Replace(s, f, x, -1), nil
}
