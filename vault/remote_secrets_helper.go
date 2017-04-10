package vault

import (
	"sync"

	"time"

	"fmt"

	log "github.com/Sirupsen/logrus"
	api "github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
	"github.com/seatgeek/hashi-helper/support"
)

func indexRemoteSecrets(environment string) SecretList {
	log.Info("Scanning for remote secrets")

	// Create a WaitGroup so we automatically unblock when all tasks are done
	var indexerWg sync.WaitGroup
	indexerCh := make(chan string, config.DefaultConcurrency*2)

	var resultWg sync.WaitGroup
	resultCh := make(chan string, config.DefaultConcurrency*2)

	// channel for signaling our go routines should stop
	completeCh := make(chan interface{})
	defer close(completeCh)

	var paths SecretList

	// Queue our first path to kick off the scanning
	indexerWg.Add(1)
	indexerCh <- "/"

	// Start go routines for workers
	for i := 0; i <= config.DefaultConcurrency; i++ {
		go remoteSecretIndexer(indexerCh, resultCh, completeCh, &indexerWg, &resultWg, i)
	}

	go remoteSecretIndexerResultProcessor(&paths, resultCh, completeCh, &resultWg)

	// Wait for all indexers to finish up
	if support.WaitTimeout(&indexerWg, time.Minute*5) {
		log.Fatal("Timeout reached (5m) waiting for remote secret scanner to complete")
	}

	// Wait for all readers to finish up
	if support.WaitTimeout(&resultWg, time.Minute*1) {
		log.Fatal("Timeout reached (5m) waiting for remote secret scanner to complete")
	}

	paths = filterByEnvironment(paths, environment)

	log.Infof("Scanning complete, found %d secrets", len(paths))

	return paths
}

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
	for i := 0; i <= config.DefaultConcurrency; i++ {
		go remoteSecretReader(readChan, completeCh, &readerWg, i)
	}

	// queue secrets to be read
	for _, secret := range secrets {
		readChan <- secret
	}

	// Wait for all remote secrets had been read (max 5m)
	if support.WaitTimeout(&readerWg, time.Minute*5) {
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

func extraEnvironmentFromPath(path string) (string, string, string, error) {
	match := environmentMatch.FindStringSubmatch(path)

	if len(match) != 4 {
		return "", "", "", fmt.Errorf("Could not parse environment from string")
	}

	return match[1], match[2], match[3], nil
}
