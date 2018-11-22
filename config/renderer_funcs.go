package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

var now = func() time.Time { return time.Now().UTC() }

// lookupVarFunc will return the template variable identified by `key` or return an error
// which will abort the template rendering.
func (r *renderer) lookupVarFunc(key string) (interface{}, error) {
	val, ok := r.variables[key]
	if !ok {
		return "", fmt.Errorf("Missing template variable '%s'", key)
	}

	return val, nil
}

// lookupVarDefaultFunc will return the template variable identified by `key` or a default value
// provided in `def`.
func (r *renderer) lookupVarDefaultFunc(key string, def interface{}) (interface{}, error) {
	val, ok := r.variables[key]
	if !ok {
		return def, nil
	}

	return val, nil
}

// lookupVarMapFunc will return the value of "mapKey" within the template variable
// identified by "key"`.
//
// If "key" is not a template variable, an error will be returned
// If "key" is not a map[string]interface{}, an error will be returned
// if "mapKey" do not exist in the map of "key", an error will be returnedd
func (r *renderer) lookupVarMapFunc(key, mapKey string) (interface{}, error) {
	if r.variablesScratch == nil {
		r.variablesScratch = &Scratch{values: r.variables}
	}

	return r.variablesScratch.MapGet(key, mapKey)
}

// lookupVarMapDefaultFunc will return the value of "mapKey" within the template variable
// identified by "key"`.
//
// If "key" is not a template variable, an error will be returned
// If "key" is not a map[string]interface{}, an error will be returned
// if "mapKey" do not exist in the map of "key", the default value provided in "def" is returned.
func (r *renderer) lookupVarMapDefaultFunc(key, mapKey string, def interface{}) (interface{}, error) {
	v, err := r.lookupVarMapFunc(key, mapKey)
	if err != nil {
		return def, nil
	}

	return v, nil
}

// consulDomainFunc will return the Consul DNS Domain.
// It will default to "consul" unless template variable key "consul_domain" is defined
func (r *renderer) consulDomainFunc() (interface{}, error) {
	return r.lookupVarDefaultFunc("consul_domain", "consul")
}

// consulServiceFunc will return a Consul Service hostname
func (r *renderer) consulServiceFunc(service string) (interface{}, error) {
	return fmt.Sprintf(`%s.service.[[ consulDomain ]]`, service), nil
}

// consulService will return a Consul Service with provided tag
func (r *renderer) consulServiceWithTagFunc(service, tag string) (interface{}, error) {
	return fmt.Sprintf(`%s.%s.service.[[ consulDomain ]]`, tag, service), nil
}

func (r *renderer) grantCredentialsFunc(db, role string) (interface{}, error) {
	tmpl := `
path "%s/creds/%s" {
  capabilities = ["read"]
}`

	return fmt.Sprintf(tmpl, db, role), nil
}

func (r *renderer) grantCredentialsPolicyFunc(db, role string) (interface{}, error) {
	tmpl := `
policy "%s-%s" {
	[[ grantCredentials "%s" "%s" ]]
}`

	return fmt.Sprintf(tmpl, db, role, db, role), nil
}

func (r *renderer) githubAssignTeamPolicyFunc(team, policy string) (interface{}, error) {
	tmpl := `
secret "/auth/github/map/teams/%s" {
  value = "%s"
}`

	return fmt.Sprintf(tmpl, team, policy), nil
}

func (r *renderer) ldapAssignTeamPolicyFunc(group, policy string) (interface{}, error) {
	tmpl := `
secret "/auth/ldap/groups/%s" {
  value = "%s"
}`

	return fmt.Sprintf(tmpl, group, policy), nil
}

// replaceAllFunc replaces all occurrences of a value in a string with the given
// replacement value.
func (r *renderer) replaceAllFunc(f, x, s string) (string, error) {
	return strings.Replace(s, f, x, -1), nil
}

// regexReplaceAllFunc replaces all occurrences of a regular expression with
// the given replacement value.
func (r *renderer) regexReplaceAllFunc(re, pl, s string) (string, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return "", err
	}
	return compiled.ReplaceAllString(s, pl), nil
}

// envFunc return a key from the process environment
func (r *renderer) envFunc(key string) (string, error) {
	return os.Getenv(key), nil
}

// base64DecodeFunc decodes the given string as a base64 string, returning an error
// if it fails.
func (r *renderer) base64DecodeFunc(s string) (string, error) {
	v, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", errors.Wrap(err, "base64Decode")
	}
	return string(v), nil
}

// base64EncodeFunc encodes the given value into a string represented as base64.
func (r *renderer) base64EncodeFunc(s string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

// base64URLDecodeFunc decodes the given string as a URL-safe base64 string.
func (r *renderer) base64URLDecodeFunc(s string) (string, error) {
	v, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", errors.Wrap(err, "base64URLDecode")
	}
	return string(v), nil
}

// base64URLEncodeFunc encodes the given string to be URL-safe.
func (r *renderer) base64URLEncodeFunc(s string) (string, error) {
	return base64.URLEncoding.EncodeToString([]byte(s)), nil
}

// containsFunc is a function that have reverse arguments of "in" and is designed to
// be used as a pipe instead of a function:
//
// 		{{ l | containsFunc "thing" }}
//
func (r *renderer) containsFunc(v, l interface{}) (bool, error) {
	return r.in(l, v)
}

// containsSomeFunc returns functions to implement each of the following:
//
// 1. containsAll    - true if (∀x ∈ v then x ∈ l); false otherwise
// 2. containsAny    - true if (∃x ∈ v such that x ∈ l); false otherwise
// 3. containsNone   - true if (∀x ∈ v then x ∉ l); false otherwise
// 2. containsNotAll - true if (∃x ∈ v such that x ∉ l); false otherwise
//
// ret_true - return true at end of loop for none/all; false for any/notall
// invert   - invert block test for all/notall
func (r *renderer) containsSomeFunc(retTrue, invert bool) func([]interface{}, interface{}) (bool, error) {
	return func(v []interface{}, l interface{}) (bool, error) {
		for i := 0; i < len(v); i++ {
			if ok, _ := r.in(l, v[i]); ok != invert {
				return !retTrue, nil
			}
		}
		return retTrue, nil
	}
}

// in searches for a given value in a given interface.
func (r *renderer) in(l, v interface{}) (bool, error) {
	lv := reflect.ValueOf(l)
	vv := reflect.ValueOf(v)

	switch lv.Kind() {
	case reflect.Array, reflect.Slice:
		// if the slice contains 'interface' elements, then the element needs to be extracted directly to examine its type,
		// otherwise it will just resolve to 'interface'.
		var interfaceSlice []interface{}
		if reflect.TypeOf(l).Elem().Kind() == reflect.Interface {
			interfaceSlice = l.([]interface{})
		}

		for i := 0; i < lv.Len(); i++ {
			var lvv reflect.Value
			if interfaceSlice != nil {
				lvv = reflect.ValueOf(interfaceSlice[i])
			} else {
				lvv = lv.Index(i)
			}

			switch lvv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch vv.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if vv.Int() == lvv.Int() {
						return true, nil
					}
				}
			case reflect.Float32, reflect.Float64:
				switch vv.Kind() {
				case reflect.Float32, reflect.Float64:
					if vv.Float() == lvv.Float() {
						return true, nil
					}
				}
			case reflect.String:
				if vv.Type() == lvv.Type() && vv.String() == lvv.String() {
					return true, nil
				}
			}
		}
	case reflect.String:
		if vv.Type() == lv.Type() && strings.Contains(lv.String(), vv.String()) {
			return true, nil
		}
	}

	return false, nil
}

// joinFunc is a version of strings.Join that can be piped
func (r *renderer) joinFunc(sep string, a []string) (string, error) {
	return strings.Join(a, sep), nil
}

// TrimSpace is a version of strings.TrimSpace that can be piped
func (r *renderer) trimSpaceFunc(s string) (string, error) {
	return strings.TrimSpace(s), nil
}

// parseBoolFunc parses a string into a boolean
func (r *renderer) parseBoolFunc(s string) (bool, error) {
	if s == "" {
		return false, nil
	}

	result, err := strconv.ParseBool(s)
	if err != nil {
		return false, errors.Wrap(err, "parseBool")
	}
	return result, nil
}

// parseFloatFunc parses a string into a base 10 float
func (r *renderer) parseFloatFunc(s string) (float64, error) {
	if s == "" {
		return 0.0, nil
	}

	result, err := strconv.ParseFloat(s, 10)
	if err != nil {
		return 0, errors.Wrap(err, "parseFloat")
	}
	return result, nil
}

// parseIntFunc parses a string into a base 10 int
func (r *renderer) parseIntFunc(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}

	result, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "parseInt")
	}
	return result, nil
}

// parseJSONFunc returns a structure for valid JSON
func (r *renderer) parseJSONFunc(s string) (interface{}, error) {
	if s == "" {
		return map[string]interface{}{}, nil
	}

	var data interface{}
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// parseUintFunc parses a string into a base 10 int
func (r *renderer) parseUintFunc(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}

	result, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "parseUint")
	}
	return result, nil
}

// pluginFunc executes a subprocess as the given command string. It is assumed the
// resulting command returns JSON which is then parsed and returned as the
// value for use in the template.
func (r *renderer) pluginFunc(name string, args ...string) (string, error) {
	if name == "" {
		return "", nil
	}

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	// Strip and trim each arg or else some plugins get confused with the newline
	// characters
	jsons := make([]string, 0, len(args))
	for _, arg := range args {
		if v := strings.TrimSpace(arg); v != "" {
			jsons = append(jsons, v)
		}
	}

	cmd := exec.Command(name, jsons...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("exec %q: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s",
			name, err, stdout.Bytes(), stderr.Bytes())
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(30 * time.Second):
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				return "", fmt.Errorf("exec %q: failed to kill", name)
			}
		}
		<-done // Allow the goroutine to exit
		return "", fmt.Errorf("exec %q: did not finish in 30s", name)
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("exec %q: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s",
				name, err, stdout.Bytes(), stderr.Bytes())
		}
	}

	return strings.TrimSpace(stdout.String()), nil
}

// regexMatchFunc returns true or false if the string matches
// the given regular expression
func (r *renderer) regexMatchFunc(re, s string) (bool, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return false, err
	}
	return compiled.MatchString(s), nil
}

// splitFunc is a version of strings.Split that can be piped
func (r *renderer) splitFunc(sep, s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}, nil
	}
	return strings.Split(s, sep), nil
}

// timestampFunc returns the current UNIX timestampFunc in UTC. If an argument is
// specified, it will be used to format the timestampFunc.
func (r *renderer) timestampFunc(s ...string) (string, error) {
	switch len(s) {
	case 0:
		return now().Format(time.RFC3339), nil
	case 1:
		if s[0] == "unix" {
			return strconv.FormatInt(now().Unix(), 10), nil
		}
		return now().Format(s[0]), nil
	default:
		return "", fmt.Errorf("timestamp: wrong number of arguments, expected 0 or 1"+
			", but got %d", len(s))
	}
}

// toLowerFunc converts the given string (usually by a pipe) to lowercase.
func (r *renderer) toLowerFunc(s string) (string, error) {
	return strings.ToLower(s), nil
}

// toJSONFunc converts the given structure into a deeply nested JSON string.
func (r *renderer) toJSONFunc(i interface{}) (string, error) {
	result, err := json.Marshal(i)
	if err != nil {
		return "", errors.Wrap(err, "toJSON")
	}
	return string(bytes.TrimSpace(result)), err
}

// toJSONPrettyFunc converts the given structure into a deeply nested pretty JSON
// string.
func (r *renderer) toJSONPrettyFunc(m map[string]interface{}) (string, error) {
	result, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "toJSONPretty")
	}
	return string(bytes.TrimSpace(result)), err
}

// toTitleFunc converts the given string (usually by a pipe) to titlecase.
func (r *renderer) toTitleFunc(s string) (string, error) {
	return strings.Title(s), nil
}

// toUpperFunc converts the given string (usually by a pipe) to uppercase.
func (r *renderer) toUpperFunc(s string) (string, error) {
	return strings.ToUpper(s), nil
}

// toYAMLFunc converts the given structure into a deeply nested YAML string.
func (r *renderer) toYAMLFunc(m map[string]interface{}) (string, error) {
	result, err := yaml.Marshal(m)
	if err != nil {
		return "", errors.Wrap(err, "toYAML")
	}
	return string(bytes.TrimSpace(result)), nil
}
