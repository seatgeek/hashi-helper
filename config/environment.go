package config

// Environment struct
// contains a list of applications found in config file(s)
type Environment struct {
	Applications map[string]Application
	// Policies     map[string]Policy
}

func (currentEnvironment Environment) merge(newEnv Environment) {
	for applicationName, application := range newEnv.Applications {
		if _, ok := currentEnvironment.Applications[applicationName]; !ok {
			currentEnvironment.Applications[applicationName] = application
		} else {
			currentEnvironment.Applications[applicationName].merge(application)
		}
	}
}
