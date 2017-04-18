package config

// Application struct
// contains a list of secrets found in config file(s)
type Application struct {
	Secrets Secrets
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
