package vault

import (
	"regexp"

	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
)

var environmentMatch = regexp.MustCompile(`^secret/(?P<Environment>.*?)/(?P<Application>.*?)/(?P<Path>.+)$`)

// SecretWriter ...
type SecretWriter interface {
	writeSecret(secret config.Secret) error
	writeEnvironment(name string, environment config.Environment) error
	getClient() *api.Client
}
