package config

// Environments struct
// contain a entry for each Environment found in config file(s)
type Environments map[string]Environment

// Environment struct
// contains a list of applications found in config file(s)
type Environment map[string]Application

// Application struct
// contains a list of secrets found in config file(s)
type Application struct {
	Secrets Secrets
}

// Secret struct
type Secret map[string]string

// Secrets struct
type Secrets map[string]Secret

func (currentEnvironments Environments) merge(newEnvs Environments) {
	for environmentName, environment := range newEnvs {
		if _, ok := currentEnvironments[environmentName]; !ok {
			currentEnvironments[environmentName] = environment
			continue
		}

		currentEnvironments[environmentName].merge(environment)
	}
}

func (currentEnvironment Environment) merge(newEnv Environment) {
	for applicationName, application := range newEnv {
		if _, ok := currentEnvironment[applicationName]; !ok {
			currentEnvironment[applicationName] = application
			continue
		}

		currentEnvironment[applicationName].merge(application)
	}
}

func (currentApplication Application) merge(newApp Application) {
	currentApplicationSecrets := currentApplication.Secrets

	for secretName, secret := range newApp.Secrets {
		if _, ok := currentApplicationSecrets[secretName]; !ok {
			currentApplicationSecrets[secretName] = secret
			continue
		}

		currentApplicationSecrets[secretName].merge(secret)
	}
}

func (currentSecrets Secrets) merge(newSecrets Secrets) {
	for secretKey, secretValue := range newSecrets {
		if _, ok := currentSecrets[secretKey]; !ok {
			currentSecrets[secretKey] = secretValue
			continue
		}

		currentSecrets[secretKey].merge(secretValue)
	}
}

func (currentSecret Secret) merge(newSecret Secret) {
	for k, v := range newSecret {
		currentSecret[k] = v
	}
}
