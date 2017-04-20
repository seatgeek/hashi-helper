package config

// Application struct
//
// environment
type Application struct {
	Secrets  Secrets
	Policies Policies
}

func (a Application) merge(newApp Application) {
	currentApplicationSecrets := a.Secrets

	for secretName, secret := range newApp.Secrets {
		if _, ok := currentApplicationSecrets[secretName]; !ok {
			currentApplicationSecrets[secretName] = secret
			continue
		}

		currentApplicationSecrets[secretName].merge(secret)
	}
}

// Applications struct
type Applications map[string]Application
