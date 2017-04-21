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
	return s.Application.Equal(o.Application) && s.Path == o.Path && s.Key == o.Key
}

// Secrets struct
//
// environment -> application
type Secrets []*Secret

// Add ...
func (e *Secrets) Add(secret *Secret) bool {
	if !e.Exists(secret) {
		*e = append(*e, secret)
		return true
	}

	return false
}

// Exists ...
func (e *Secrets) Exists(secret *Secret) bool {
	for _, existing := range *e {
		if secret.Equal(existing) {
			return true
		}
	}

	return false
}

// Get ...
func (e *Secrets) Get(secret *Secret) *Secret {
	for _, existing := range *e {
		if secret.Equal(existing) {
			return existing
		}
	}

	return nil
}

// GetOrSet ...
func (e *Secrets) GetOrSet(secret *Secret) *Secret {
	existing := e.Get(secret)
	if existing != nil {
		return existing
	}

	e.Add(secret)
	return secret
}
