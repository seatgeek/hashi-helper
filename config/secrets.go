package config

// Secrets struct
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
