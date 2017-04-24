package helper

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
)

// SecretWriter ...
type SecretWriter struct {
	client *api.Client
}

// WriteSecret ...
func (w SecretWriter) WriteSecret(secret *config.Secret) error {
	path := fmt.Sprintf("secret/%s/%s", secret.Application.Name, secret.Path)
	log.Info(path)

	_, err := w.getClient().Logical().Write(path, secret.Secret.Data)
	return err
}

func (w SecretWriter) getClient() *api.Client {
	if w.client == nil {
		client, err := api.NewClient(nil)
		if err != nil {
			log.Fatal(err)
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
