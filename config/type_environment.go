package config

import "fmt"

// Environment struct
type Environment struct {
	Name         string
	Applications Applications
	Policies     Policies
	Mounts       Mounts
}

func (e Environment) merge(newEnv Environment) {
	for applicationName, application := range newEnv.Applications {
		if _, ok := e.Applications[applicationName]; !ok {
			e.Applications[applicationName] = application
		} else {
			e.Applications[applicationName].merge(application)
		}
	}
}

// Environments struct
type Environments map[string]Environment

// Get ...
func (e Environments) Get(env string) (Environment, error) {
	if val, ok := e[env]; ok {
		return val, nil
	}

	return Environment{}, fmt.Errorf("Could not find environment %s", env)
}

// Contains ...
func (e Environments) Contains(env string) bool {
	if _, ok := e[env]; ok {
		return ok
	}
	return false
}

func (e Environments) merge(newEnvs Environments) {
	for environmentName, environment := range newEnvs {
		if _, ok := e[environmentName]; !ok {
			e[environmentName] = environment
			continue
		}

		e[environmentName].merge(environment)
	}
}
