package config

// Application struct
//
// environment
type Application struct {
	Secrets  Secrets
	Policies Policies
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
