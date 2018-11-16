package config

import (
	"errors"
	"fmt"
)

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
