package vault

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
)

// SharedSecretWriter ...
type SharedSecretWriter struct {
	client *api.Client
}

func (w SharedSecretWriter) writeEnvironment(name string, e config.Environment) error {
	for _, application := range e.Applications {
		err := w.writeApplication(application)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w SharedSecretWriter) writeApplication(application config.Application) error {
	for _, secret := range application.Secrets {
		err := w.writeSecret(secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w SharedSecretWriter) writeSecret(secret config.Secret) error {
	path := fmt.Sprintf("secret/%s/%s/%s", secret.Environment, secret.Application, secret.Path)
	log.Info(path)

	_, err := w.getClient().Logical().Write(path, secret.Secret.Data)
	return err
}

func (w SharedSecretWriter) getClient() *api.Client {
	if w.client == nil {
		client, err := api.NewClient(nil)
		if err != nil {
			log.Fatal(err)
		}
		w.client = client
	}
	return w.client
}
