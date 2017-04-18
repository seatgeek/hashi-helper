package config

import "fmt"

// Environments struct
// contain a entry for each Environment found in config file(s)
type Environments map[string]Environment

func (e Environments) Get(env string) (Environment, error) {
	if val, ok := e[env]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("Could not find environment %s", env)
}

// Contains ...
func (e Environments) Contains(env string) bool {
	if _, ok := e[env]; ok {
		return ok
	}
	return false
}

func (currentEnvironments Environments) merge(newEnvs Environments) {
	for environmentName, environment := range newEnvs {
		if _, ok := currentEnvironments[environmentName]; !ok {
			currentEnvironments[environmentName] = environment
			continue
		}

		currentEnvironments[environmentName].merge(environment)
	}
}
