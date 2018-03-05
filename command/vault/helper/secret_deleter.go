package helper

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
)

// SecretDeleter ...
type SecretDeleter struct {
	client *api.Client
}

// DeleteSecret ...
func (w SecretDeleter) DeleteSecret(secret *config.Secret, env string) error {
	var path string

	// @TODO Make a dedicated type for writing non-secrets !
	if strings.HasPrefix(secret.Path, "/") {
		path = strings.TrimLeft(secret.Path, "/")
	} else if secret.Application != nil {
		path = fmt.Sprintf("secret/%s/%s/%s", env, secret.Application.Name, secret.Path)
	} else {
		path = fmt.Sprintf("secret/%s", secret.Path)
	}

	log.Warnf("Deleting secret: %s", path)

	_, err := w.getClient().Logical().Delete(path)
	return err
}

func (w SecretDeleter) getClient() *api.Client {
	if w.client == nil {
		client, err := api.NewClient(nil)
		if err != nil {
			log.Fatal(err)
		}
		w.client = client
	}
	return w.client
}
