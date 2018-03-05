package vault

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/vault/api"
	cli "gopkg.in/urfave/cli.v1"
)

// UnsealKeybase ...
func UnsealKeybase(c *cli.Context) error {
	str := c.String("unseal-key")
	if str == "" {
		return fmt.Errorf("missing VAULT_UNSEAL_KEY env or --unseal--key cli flag")
	}

	log.Info("Decoding VAULT_UNSEAL_KEY")
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return fmt.Errorf("Could not base64decode the unseal key")
	}

	log.Info("Decrypting VAULT_UNSEAL_KEY")
	cmd := exec.Command("keybase", "pgp", "decrypt")
	cmd.Stdin = strings.NewReader(string(data))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run keybase gpg decrypt: %s - %s", err, stderr.String())
	}
	token := stdout.String()

	service := c.String("consul-service-name")
	if service == "" { // assume VAULT_ADDR set
		return sendUnseal(token, nil)
	}

	// assume we should unseal multiple clients
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		return err
	}

	services, _, err := client.Catalog().Service(service, c.String("consul-service-tag"), nil)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return fmt.Errorf("Could not find any vault instances in consul service '%s' with tag '%s'", service, c.String("consul-service-tag"))
	}

	log.Infof("Found %d vault instances in consul service '%s' with tag '%s'", len(services), service, c.String("consul-service-tag"))

	for _, service := range services {
		cfg := api.DefaultConfig()
		cfg.ReadEnvironment()
		cfg.Address = fmt.Sprintf("%s://%s:%d", c.String("vault-protocol"), service.ServiceAddress, service.ServicePort)

		err := sendUnseal(token, cfg)
		if err != nil {
			log.Error(err)
			continue
		}
	}

	return nil
}

func sendUnseal(token string, config *api.Config) error {
	var logger *log.Entry

	if config == nil || config.Address == "" {
		logger = log.WithField("vault-server", "${VAULT_ADDR}")
	} else {
		logger = log.WithField("vault-server", config.Address)
	}

	logger.Info("Sending unseal command to Vault")
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	resp, err := client.Sys().Unseal(token)
	if err != nil {
		return fmt.Errorf("Could not unseal Vault: %s", err)
	}

	logger.Info("Unseal succeeded")

	if !resp.Sealed {
		logger.Info("Vault instance is now unsealed")
	} else {
		logger.Warnf("Vault unseal progress: %d out of %d unseal keys has been provided (%d shares)", resp.Progress, resp.T, resp.N)
	}

	return nil
}
