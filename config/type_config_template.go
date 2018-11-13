package config

import (
	"errors"
	"fmt"
)

func (c *Config) consulDomain() (string, error) {
	val, ok := c.Interpolations["consul_domain"]
	if !ok {
		return "", errors.New("Missing interpolation key 'consul_domain'")
	}

	return val, nil
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
