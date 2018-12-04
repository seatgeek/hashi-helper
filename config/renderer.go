package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type renderer struct {
	variables        map[string]interface{}
	variablesScratch *Scratch
	scratch          *Scratch
	readConfigFiles  []string
}

func newRenderer(variables, variableFiles []string) (*renderer, error) {
	r := &renderer{
		variables: map[string]interface{}{},
		scratch:   &Scratch{},
	}

	if err := r.readTemplateVariablesFiles(variableFiles); err != nil {
		return nil, err
	}

	if err := r.parseTemplateVariables(variables); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *renderer) renderContent(content, file string, depth int) (string, error) {
	log.Debugf("Rendering file %s (depth %d)", file, depth)

	if depth > 10 {
		return "", fmt.Errorf("recursive template rendering found, aborting")
	}

	fns := template.FuncMap{
		"base64Decode":           r.base64DecodeFunc,
		"base64Encode":           r.base64EncodeFunc,
		"base64URLDecode":        r.base64URLDecodeFunc,
		"base64URLEncode":        r.base64URLEncodeFunc,
		"consulDomain":           r.consulDomainFunc,
		"contains":               r.containsFunc,
		"containsAll":            r.containsSomeFunc(true, true),
		"containsAny":            r.containsSomeFunc(false, false),
		"containsNone":           r.containsSomeFunc(true, false),
		"containsNotAll":         r.containsSomeFunc(false, true),
		"env":                    r.envFunc,
		"githubAssignTeamPolicy": r.githubAssignTeamPolicyFunc,
		"grantCredentials":       r.grantCredentialsFunc,
		"grantCredentialsPolicy": r.grantCredentialsPolicyFunc,
		"in":                     r.in,
		"join":                   r.joinFunc,
		"ldapAssignGroupPolicy":  r.ldapAssignTeamPolicyFunc,
		"lookup":                 r.lookupVarFunc,
		"lookupDefault":          r.lookupVarDefaultFunc,
		"lookupMap":              r.lookupVarMapFunc,
		"lookupMapDefault":       r.lookupVarMapDefaultFunc,
		"parseBool":              r.parseBoolFunc,
		"parseFloat":             r.parseFloatFunc,
		"parseInt":               r.parseIntFunc,
		"parseJSON":              r.parseJSONFunc,
		"parseUint":              r.parseUintFunc,
		"plugin":                 r.pluginFunc,
		"regexMatch":             r.regexMatchFunc,
		"regexReplaceAll":        r.regexReplaceAllFunc,
		"replaceAll":             r.replaceAllFunc,
		"scratch":                r.createScratch(),
		"service":                r.consulServiceFunc,
		"serviceWithTag":         r.consulServiceWithTagFunc,
		"split":                  r.splitFunc,
		"timestamp":              r.timestampFunc,
		"toJSON":                 r.toJSONFunc,
		"toJSONPretty":           r.toJSONPrettyFunc,
		"toLower":                r.toLowerFunc,
		"toTitle":                r.toTitleFunc,
		"toUpper":                r.toUpperFunc,
		"toYAML":                 r.toYAMLFunc,
		"trimSpace":              r.trimSpaceFunc,
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
	if err := tmpl.Execute(writer, r.variables); err != nil {
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
		return r.renderContent(content, file, depth+1)
	}

	// HCL pretty print the rendered file
	res, err := printer.Format(b.Bytes())
	if err != nil {
		return "", fmt.Errorf("Could not format HCL file %s: %s", file, err)
	}

	// Trim the string for spaces / newlines and return the result
	return strings.TrimSpace(string(res)), nil
}

func (r *renderer) parseTemplateVariables(pairs []string) error {
	for _, val := range pairs {
		chunks := strings.SplitN(val, "=", 2)
		if len(chunks) != 2 {
			return fmt.Errorf("Interpolation key/value pair '%s' is not valid", val)
		}

		r.variables[chunks[0]] = chunks[1]
	}

	return nil
}

func (r *renderer) readTemplateVariablesFiles(files []string) error {
	for _, variableFile := range files {
		ext := path.Ext(variableFile)

		var variables map[string]interface{}
		var err error

		switch ext {
		case ".hcl":
			variables, err = r.parseHCLVars(variableFile)
		case ".yaml", ".yml":
			variables, err = r.parseYAMLVars(variableFile)
		case ".json":
			variables, err = r.parseJSONVars(variableFile)
		default:
			err = fmt.Errorf("variables file extension %v not supported", ext)
		}

		if err != nil {
			return err
		}

		read, _ := filepath.Abs(variableFile)
		r.readConfigFiles = append(r.readConfigFiles, read)

		for k, v := range variables {
			r.variables[k] = v
		}
	}

	return nil
}

// parseJSONVars will read a file from disk and JSON unmarshal it into a map[string]interface{}
func (r *renderer) parseJSONVars(variableFile string) (variables map[string]interface{}, err error) {
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
func (r *renderer) parseYAMLVars(variableFile string) (variables map[string]interface{}, err error) {
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
func (r *renderer) parseHCLVars(variableFile string) (variables map[string]interface{}, err error) {
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

func (r *renderer) createScratch() func() *Scratch {
	return func() *Scratch {
		if r.scratch == nil {
			r.scratch = &Scratch{}
		}
		return r.scratch
	}
}
