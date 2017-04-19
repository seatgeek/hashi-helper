package config

// Secrets struct
//
// environment -> application
type Secrets map[string]InternalSecret

func (currentSecrets Secrets) merge(newSecrets Secrets) {
	for secretKey, secretValue := range newSecrets {
		if _, ok := currentSecrets[secretKey]; !ok {
			currentSecrets[secretKey] = secretValue
			continue
		}

		currentSecrets[secretKey].merge(secretValue)
	}
}
