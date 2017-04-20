package config

import (
	"sort"

	"github.com/hashicorp/vault/api"
)

// Secret ...
type Secret struct {
	Path        string
	Environment string
	Application string
	Key         string
	Secret      *api.Secret
}

func (s Secret) merge(newSecret Secret) {
	for k, v := range newSecret.Secret.Data {
		s.Secret.Data[k] = v
	}
}

// Secrets struct
//
// environment -> application
type Secrets map[string]Secret

func (currentSecrets Secrets) merge(newSecrets Secrets) {
	for secretKey, secretValue := range newSecrets {
		if _, ok := currentSecrets[secretKey]; !ok {
			currentSecrets[secretKey] = secretValue
			continue
		}

		currentSecrets[secretKey].merge(secretValue)
	}
}

// SecretList ...
type SecretList []*Secret

func (p SecretList) Len() int { return len(p) }
func (p SecretList) Less(i, j int) bool {
	return p[i].Environment+"_"+p[i].Application+"_"+p[i].Path < p[j].Environment+"_"+p[j].Application+"_"+p[j].Path
}
func (p SecretList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p SecretList) Sort() { sort.Sort(p) }
