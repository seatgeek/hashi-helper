package main

import (
	"sync"

	"time"

	log "github.com/Sirupsen/logrus"
	api "github.com/hashicorp/vault/api"
)

// InternalSecret ...
type InternalSecret struct {
	Path string
	*api.Secret
}

// SecretList ...
type SecretList map[string]*InternalSecret

// readRemoteSecrets
// Take an array of secret paths to read
func readRemoteSecrets(secrets []string) (SecretList, error) {
	log.Infof("Going to read %d remote secrets", len(secrets))

	// Create a WaitGroup for the remote reader, so we automatically unblock when all tasks are done
	var readerWg sync.WaitGroup
	readerWg.Add(len(secrets))

	// Create a WaitGroup for the result processor, so we automatically unblock when all tasks are done
	var resultWg sync.WaitGroup
	resultWg.Add(len(secrets))

	// Signal processing is done
	completeCh := make(chan interface{})
	defer close(completeCh)

	// Vault paths to be read
	readChan := make(chan string, defaultConcurrency*2)
	defer close(readChan)

	// Vault secrets read from remote
	resultCh := make(chan *InternalSecret, defaultConcurrency*2)
	defer close(resultCh)

	// Result from processor will be stored here
	secretResult := make(SecretList, len(secrets))

	// Start go routines for readers
	for i := 0; i <= defaultConcurrency; i++ {
		go remoteSecretReader(readChan, resultCh, completeCh, &readerWg, i)
	}

	// Convert the reader result to InternalSecret
	go remoteSecretResultProcessor(secretResult, resultCh, completeCh, &resultWg)

	// queue secrets to be read
	for _, secret := range secrets {
		readChan <- secret
	}

	// Wait for all remote secrets had been read (max 5m)
	if waitTimeout(&readerWg, time.Minute*5) {
		log.Fatal("Timeout reached (5m) waiting for remote reader to complete")
	}
	log.Info("Remote secret reader complete")

	// Wait for results to have been processed (max 1m)
	if waitTimeout(&resultWg, time.Minute*1) {
		log.Fatal("Timeout reached (1m) waiting for result reader to complete")
	}
	log.Info("Result reader complete")

	log.Infof("Number of secrets read: %d", len(secretResult))

	// Return the secrets
	return secretResult, nil
}

func remoteSecretResultProcessor(result SecretList, resultCh chan *InternalSecret, completeCh chan interface{}, wg *sync.WaitGroup) {
	for {
		select {
		case <-completeCh:
			return
		case secret := <-resultCh:
			result[secret.Path] = secret
			wg.Done()
		}
	}
}

func remoteSecretReader(readCh chan string, resultCh chan *InternalSecret, completeCh chan interface{}, wg *sync.WaitGroup, workerID int) error {
	log.WithField("method", "remoteSecretFetcher").Debugf("Starting worker %d", workerID)

	// Create a new Vault API client for this go-routine
	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-completeCh:
			log.WithField("method", "remoteSecretIndexer").Debugf("Stopping worker %d", workerID)
			return nil
		case path := <-readCh:
			log.WithField("method", "remoteSecretIndexer").WithField("path", path).Debugf("[%d] Reading secret", workerID)
			secret, err := client.Logical().Read(path)
			if err != nil {
				return err
			}

			resultCh <- &InternalSecret{path, secret}
			wg.Done()
		}
	}
}
