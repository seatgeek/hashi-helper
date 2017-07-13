package vault

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
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

	log.Info("Sending unseal command to Vault")
	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	resp, err := client.Sys().Unseal(token)
	if err != nil {
		return fmt.Errorf("Could not unseal Vault: %s", err)
	}

	log.Info("Unseal succeeded")

	if !resp.Sealed {
		log.Info("Vault instance is now unsealed")
	} else {
		log.Warnf("Vault unseal progress: %d out of %d unseal keys has been provided (%d shares)", resp.Progress, resp.T, resp.N)
	}

	return nil
}
