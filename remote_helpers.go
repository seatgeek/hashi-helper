package main

import (
	"regexp"
	"sync"

	"time"

	"fmt"

	log "github.com/Sirupsen/logrus"
	api "github.com/hashicorp/vault/api"
)

var environmentMatch = regexp.MustCompile(`^secret/(?P<Environment>.*?)/(?P<Application>.*?)/(?P<Path>.+)$`)

// InternalSecret ...
type InternalSecret struct {
	Path        string
	Environment string
	Application string
	Secret      *api.Secret
}

// SecretList ...
type SecretList []*InternalSecret

// readRemoteSecrets
// Take an array of secret paths to read
func readRemoteSecrets(secrets SecretList) (SecretList, error) {
	log.Infof("Going to read %d remote secrets", len(secrets))

	// Create a WaitGroup for the remote reader, so we automatically unblock when all tasks are done
	var readerWg sync.WaitGroup
	readerWg.Add(len(secrets))

	// Signal processing is done
	completeCh := make(chan interface{})
	defer func() { close(completeCh) }()

	// Vault paths to be read
	readChan := make(chan *InternalSecret, len(secrets))

	// Start go routines for readers
	for i := 0; i <= defaultConcurrency; i++ {
		go remoteSecretReader(readChan, completeCh, &readerWg, i)
	}

	// queue secrets to be read
	for _, secret := range secrets {
		readChan <- secret
	}

	// Wait for all remote secrets had been read (max 5m)
	if waitTimeout(&readerWg, time.Minute*5) {
		log.Fatal("Timeout reached (5m) waiting for remote reader to complete")
	}
	log.Info("Remote secret reader complete")

	log.Infof("Number of secrets read: %d", len(secrets))

	// Return the secrets
	return secrets, nil
}

func remoteSecretReader(readCh chan *InternalSecret, completeCh chan interface{}, wg *sync.WaitGroup, workerID int) error {
	log.WithField("method", "remoteSecretFetcher").Debugf("Starting worker %d", workerID)

	// Create a new Vault API client for this go-routine
	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-completeCh:
			log.Debugf("Stopping worker %d", workerID)
			return nil
		case secret := <-readCh:
			log.Debugf("[%d] Reading secret %s", workerID, secret.Path)
			remoteSecret, err := client.Logical().Read(secret.Path)
			if err != nil {
				log.Fatal(err)
			}

			secret.Secret = remoteSecret
			wg.Done()
		}
	}
}

func extraEnvironmentFromPath(path string) (string, string, error) {
	match := environmentMatch.FindStringSubmatch(path)

	if len(match) != 4 {
		return "", "", fmt.Errorf("Could not parse environment from string")
	}

	return match[1], match[2], nil

}
