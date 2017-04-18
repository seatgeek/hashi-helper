package config

import "github.com/hashicorp/vault/api"

// InternalSecret ...
type InternalSecret struct {
	Path        string
	Environment string
	Application string
	Key         string
	Secret      *api.Secret
}

func (currentSecret InternalSecret) merge(newSecret InternalSecret) {
	for k, v := range newSecret.Secret.Data {
		currentSecret.Secret.Data[k] = v
	}
}
