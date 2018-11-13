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
  capabilities = ["read", "list"]
}`

	return fmt.Sprintf(tmpl, db, role), nil
}
