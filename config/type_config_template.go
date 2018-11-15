package config

import (
	"errors"
	"fmt"
)

func (c *Config) consulDomain() (string, error) {
	val, ok := c.templateVariables["consul_domain"]
	if !ok {
		return "", errors.New("Missing interpolation key 'consul_domain'")
	}

	return fmt.Sprintf("%+v", val), nil
}

func (c *Config) service(service string) (interface{}, error) {
	domain, err := c.consulDomain()
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("%s.service.%s", service, domain), nil
}

func (c *Config) serviceWithTag(service, tag string) (interface{}, error) {
	domain, err := c.consulDomain()
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("%s.%s.service.%s", tag, service, domain), nil
}

func (c *Config) grantCredentials(db, role string) (interface{}, error) {
	tmpl := `
path "%s/creds/%s" {
  capabilities = ["read"]
}`

	return fmt.Sprintf(tmpl, db, role), nil
}

func (c *Config) grantCredentialsPolicy(db, role string) (interface{}, error) {
	tmpl := `
policy "%s-%s" {
	[[ grant_credentials "%s" "%s" ]]
}`

	return fmt.Sprintf(tmpl, db, role, db, role), nil
}

func (c *Config) githubAssignTeamPolicy(team, policy string) (interface{}, error) {
	tmpl := `
secret "/auth/github/map/teams/%s" {
  value = "%s"
}`

	return fmt.Sprintf(tmpl, team, policy), nil
}

func (c *Config) ldapAssignTeamPolicy(group, policy string) (interface{}, error) {
	tmpl := `
secret "/auth/ldap/groups/%s" {
  value = "%s"
}`

	return fmt.Sprintf(tmpl, group, policy), nil
}
