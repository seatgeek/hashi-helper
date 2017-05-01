package helper

import (
	"regexp"
	"strings"
	"sync"

	"time"

	"fmt"

	log "github.com/Sirupsen/logrus"
	api "github.com/hashicorp/vault/api"
	"github.com/seatgeek/hashi-helper/config"
	"github.com/seatgeek/hashi-helper/support"
)

var environmentMatch = regexp.MustCompile(`^secret/(?P<Environment>.*?)/(?P<Application>.*?)/(?P<Path>.+)$`)

// IndexRemoteSecrets ...
func IndexRemoteSecrets(environment string) config.VaultSecrets {
	log.Info("Scanning for remote secrets")

	// Create a WaitGroup so we automatically unblock when all tasks are done
	var indexerWg sync.WaitGroup
	indexerCh := make(chan string, config.DefaultConcurrency*2)

	var resultWg sync.WaitGroup
	resultCh := make(chan string, config.DefaultConcurrency*2)

	// channel for signaling our go routines should stop
	completeCh := make(chan interface{})
	defer close(completeCh)

	var paths config.VaultSecrets

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
func ReadRemoteSecrets(secrets config.VaultSecrets) (config.VaultSecrets, error) {
	log.Infof("Going to read %d remote secrets", len(secrets))

	// Create a WaitGroup for the remote reader, so we automatically unblock when all tasks are done
	var readerWg sync.WaitGroup
	readerWg.Add(len(secrets))

	// Signal processing is done
	completeCh := make(chan interface{})
	defer func() { close(completeCh) }()

	// Vault paths to be read
	readChan := make(chan *config.Secret, len(secrets))

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

func remoteSecretReader(readCh chan *config.Secret, completeCh chan interface{}, wg *sync.WaitGroup, workerID int) error {
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

func filterByEnvironment(secrets config.VaultSecrets, environment string) (result config.VaultSecrets) {
	if environment == "" {
		return secrets
	}

	for _, s := range secrets {
		if s.Environment.Name == environment {
			result = append(result, s)
		}
	}

	return result
}

func remoteSecretIndexerResultProcessor(result *config.VaultSecrets, resultCh chan string, completeCh chan interface{}, wg *sync.WaitGroup) {
	apps := make(map[string]*config.Application)
	envs := make(map[string]*config.Environment)

	for {
		select {
		case <-completeCh:
			return
		case path := <-resultCh:
			environment, application, key, err := extraEnvironmentFromPath(path)
			if err != nil {
				environment = "unknown"
				log.Warnf("Could not extract environment from %s", path)
			}

			if _, ok := apps[application]; !ok {
				apps[application] = &config.Application{Name: application}
			}

			if _, ok := envs[environment]; !ok {
				envs[environment] = &config.Environment{Name: environment}
			}

			*result = append(*result, &config.Secret{
				Path:        path,
				Key:         key,
				Environment: envs[environment],
				Application: apps[application],
			})
			wg.Done()
		}
	}
}

func remoteSecretIndexer(indexerCh chan string, resultCh chan string, completeCh chan interface{}, indexerWg *sync.WaitGroup, resultWg *sync.WaitGroup, workerID int) {
	log.Debugf("Starting worker %d", workerID)

	// Create a new Vault API client for this go-routine
	client, err := api.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-completeCh:
			log.Debugf("Stopping worker %d", workerID)
			return
		case path := <-indexerCh:
			logicalPath := fmt.Sprintf("secret/%s", strings.Trim(path, "/"))
			log.Debugf("[%d] Scanning path: %s", workerID, logicalPath)

			response, err := client.Logical().List(logicalPath)
			if err != nil {
				log.Fatal(err)
			}

			if response.Data == nil {
				log.Fatal("Response contains no data")
			}

			rawKeys, ok := response.Data["keys"]
			if !ok {
				log.Fatal("Could not find any keys in the response, server issues?")
			}

			keys := secretsToString(rawKeys)

			if len(keys) == 0 {
				log.Fatalf("No keys found in the vault path %s", path)
			}

			for _, keyPath := range keys {
				// If the path end in a /, it's a "directory" and should be processed recursively
				if strings.HasSuffix(keyPath, "/") {
					indexerWg.Add(1)
					indexerCh <- fmt.Sprintf("%s/%s", path, strings.Trim(keyPath, "/"))
					continue
				}

				// Add the found secret to the result
				resultWg.Add(1)
				resultCh <- fmt.Sprintf("%s/%s", logicalPath, keyPath)
			}

			indexerWg.Done()
		}
	}
}

func secretsToString(in interface{}) (out []string) {
	t := in.([]interface{})

	for _, v := range t {
		out = append(out, v.(string))
	}

	return out
}
