package config

// Environment struct
type Environment struct {
	Name         string
	Applications Applications
	Policies     Policies
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
