package config

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
