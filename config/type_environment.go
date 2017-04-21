package config

// Environment struct
type Environment struct {
	Name         string
	Applications Applications
}

// Equal ...
func (e *Environment) Equal(o *Environment) bool {
	return e.Name == o.Name
}

// Environments struct
type Environments []*Environment

// Add ...
func (e *Environments) Add(environment *Environment) {
	if !e.Exists(environment) {
		*e = append(*e, environment)
	}
}

// Exists ...
func (e *Environments) Exists(environment *Environment) bool {
	for _, existing := range *e {
		if environment.Equal(existing) {
			return true
		}
	}

	return false
}

// Get ...
func (e *Environments) Get(environment *Environment) *Environment {
	for _, existing := range *e {
		if environment.Equal(existing) {
			return existing
		}
	}

	return nil
}

// GetOrSet ...
func (e *Environments) GetOrSet(environment *Environment) *Environment {
	existing := e.Get(environment)
	if existing != nil {
		return existing
	}

	e.Add(environment)
	return environment
}
