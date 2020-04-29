package helper

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
	log "github.com/sirupsen/logrus"
)

// SecretWriter ...
type SecretWriter struct {
	client *api.Client
}

// WriteSecret ...
func (w SecretWriter) WriteSecret(secret *config.Secret, config map[string]string) error {
	var path string

	// @TODO Make a dedicated type for writing non-secrets !
	if strings.HasPrefix(secret.Path, "/") {
		path = strings.TrimLeft(secret.Path, "/")
	} else if secret.Application != nil {
		path = fmt.Sprintf("secret/%s/%s", secret.Application.Name, secret.Path)
	} else {
		path = fmt.Sprintf("secret/%s", secret.Path)
	}

	if prefix, ok := config["only-prefix"]; ok && !strings.HasPrefix(path, prefix) {
		log.Infof("Skipping %s, does not match prefix %s", path, prefix)
		return nil
	}

	log.Info(path)

	_, err := w.getClient().Logical().Write(path, secret.VaultSecret.Data)
	return err
}

func (w SecretWriter) getClient() *api.Client {
	if w.client == nil {
		client, err := api.NewClient(nil)
		if err != nil {
			log.Panic(err)
		}
		w.client = client
	}
	return w.client
}

// EscapeValue ...
func EscapeValue(value string) string {
	value = strings.Replace(value, "\\", "\\\\", -1)
	value = strings.Replace(value, "\"", "\\", -1)

	return value
}

// FormatHclFile ...
func FormatHclFile(file string) (bytes.Buffer, error) {
	cmd := exec.Command("hclfmt", "-w", file)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	return out, err
}
