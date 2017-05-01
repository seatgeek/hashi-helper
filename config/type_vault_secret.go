package config

import "github.com/hashicorp/vault/api"

// Secret ...
type Secret struct {
	secret      *Secret
	Application *Application
	Environment *Environment
	Path        string
	Key         string
	Secret      *api.Secret
}

// Equal ...
func (s *Secret) Equal(o *Secret) bool {
	if s.Application != nil && o.Application != nil {
		if s.Application.Equal(o.Application) == false {
			return false
		}
	}

	return s.Path == o.Path && s.Key == o.Key
}

// VaultSecrets struct
//
// environment -> application
type VaultSecrets []*Secret

// Add ...
func (e *VaultSecrets) Add(secret *Secret) bool {
	if !e.Exists(secret) {
		*e = append(*e, secret)
		return true
	}

	return false
}

// Exists ...
func (e *VaultSecrets) Exists(secret *Secret) bool {
	for _, existing := range *e {
		if secret.Equal(existing) {
			return true
		}
	}

	return false
}

// Get ...
func (e *VaultSecrets) Get(secret *Secret) *Secret {
	for _, existing := range *e {
		if secret.Equal(existing) {
			return existing
		}
	}

	return nil
}

// GetOrSet ...
func (e *VaultSecrets) GetOrSet(secret *Secret) *Secret {
	existing := e.Get(secret)
	if existing != nil {
		return existing
	}

	e.Add(secret)
	return secret
}
