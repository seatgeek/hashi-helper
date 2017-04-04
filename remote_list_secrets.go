package main

import (
	"fmt"
	"strings"

	"sync"

	"time"

	log "github.com/Sirupsen/logrus"
	api "github.com/hashicorp/vault/api"
	cli "gopkg.in/urfave/cli.v1"
)

func listRemoteSecretsCommand(c *cli.Context) error {
	log.Info("Scanning for remote secrets")

	// Create a WaitGroup so we automatically unblock when all tasks are done
	var indexerWg sync.WaitGroup

	var resultWg sync.WaitGroup

	// channel for signaling our go routines should stop
	completeCh := make(chan interface{})
	defer close(completeCh)

	// Channel for Vault paths to index
	readCh := make(chan string, defaultConcurrency*2)

	// Channel for Vault secrets found by indexer
	resultCh := make(chan string, defaultConcurrency*2)

	paths := make([]string, 0)

	// Start go routines for workers
	for i := 0; i <= defaultConcurrency; i++ {
		go remoteSecretIndexer(readCh, resultCh, completeCh, &indexerWg, &resultWg, i)
	}

	go remoteSecretIndexerResultProcessor(&paths, resultCh, completeCh, &resultWg)

	// Queue our first path to kick off the scanning
	indexerWg.Add(1)
	readCh <- "/"

	// Wait for all indexers to finish up
	if waitTimeout(&indexerWg, time.Minute*5) {
		log.Fatal("Timeout reached (5m) waiting for remote secret scanner to complete")
	}

	// Wait for all readers to finish up
	if waitTimeout(&resultWg, time.Minute*1) {
		log.Fatal("Timeout reached (5m) waiting for remote secret scanner to complete")
	}

	log.Infof("Scanning complete, found %d secrets", len(paths))

	readRemoteSecrets(paths)

	return nil
}

func remoteSecretIndexerResultProcessor(result *[]string, resultCh chan string, completeCh chan interface{}, wg *sync.WaitGroup) {
	for {
		select {
		case <-completeCh:
			return
		case path := <-resultCh:
			*result = append(*result, path)
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
