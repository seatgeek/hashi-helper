package config

// Environment struct
// contains a list of applications found in config file(s)
type Environment map[string]Application

func (currentEnvironment Environment) merge(newEnv Environment) {
	for applicationName, application := range newEnv {
		if _, ok := currentEnvironment[applicationName]; !ok {
			currentEnvironment[applicationName] = application
			continue
		}

		currentEnvironment[applicationName].merge(application)
	}
}
