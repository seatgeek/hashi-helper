package config

// Application ...
type Application struct {
	Name        string
	Environment *Environment
}

// Equal ...
func (a *Application) Equal(o *Application) bool {
	if a.Name != o.Name {
		return false
	}

	if !a.Environment.Equal(o.Environment) {
		return false
	}

	return true
}

// Applications ...
type Applications []*Application

// Add ...
func (a *Applications) Add(application *Application) {
	if !a.Exists(application) {
		*a = append(*a, application)
	}
}

// Exists ...
func (a *Applications) Exists(application *Application) bool {
	for _, existing := range *a {
		if application.Equal(existing) {
			return true
		}
	}

	return false
}

// Get ...
func (a *Applications) Get(application *Application) *Application {
	for _, existing := range *a {
		if application.Equal(existing) {
			return existing
		}
	}

	return nil
}

// GetOrSet ...
func (a *Applications) GetOrSet(application *Application) *Application {
	existing := a.Get(application)
	if existing != nil {
		return existing
	}

	a.Add(application)
	return application
}

func (a *Applications) List() []string {
	res := []string{}

	for _, app := range *a {
		res = append(res, app.Name)
	}

	return res
}
