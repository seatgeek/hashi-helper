package vault

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	cli "gopkg.in/urfave/cli.v1"
)

// CreateToken ...
func CreateToken(c *cli.Context) error {
	path := "auth/token/create"
	if c.Bool("orphan") {
		path = "auth/token/create-orphan"
	}

	payload := make(map[string]interface{})
	if c.String("id") != "" {
		payload["id"] = c.String("id")
	}
	if c.String("display-name") != "" {
		payload["display_name"] = c.String("display-name")
	}
	if c.String("ttl") != "" {
		payload["ttl"] = c.String("ttl")
	}
	if c.String("period") != "" {
		payload["period"] = c.String("period")
	}
	if len(c.StringSlice("policy")) > 0 {
		payload["policy"] = c.StringSlice("policy")
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	log.Info("Creating token")
	response, err := client.Logical().Write(path, payload)
	if err != nil {
		return err
	}

	log.Info("Got token")

	token := response.Auth.ClientToken
	if len(c.StringSlice("keybase")) == 0 {
		log.Info("New token: %s", token)
		return nil
	}

	log.Info("Encrypting token")

	args := make([]string, 0)
	args = append(args, "encrypt")
	args = append(args, c.StringSlice("keybase")...)

	cmd := exec.Command("keybase", args...)
	cmd.Stdin = strings.NewReader(token)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run keybase encrypt: %s - %s", err, stderr.String())
	}

	log.Infof("Send the following message to %s:", strings.Join(c.StringSlice("keybase"), ","))
	fmt.Println()
	fmt.Println(stdout.String())

	return nil
}
