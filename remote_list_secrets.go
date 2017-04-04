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
	indexerCh := make(chan string, defaultConcurrency*2)

	var resultWg sync.WaitGroup
	resultCh := make(chan string, defaultConcurrency*2)

	// channel for signaling our go routines should stop
	completeCh := make(chan interface{})
	defer close(completeCh)

	var paths SecretList

	// Queue our first path to kick off the scanning
	indexerWg.Add(1)
	indexerCh <- "/"

	// Start go routines for workers
	for i := 0; i <= defaultConcurrency; i++ {
		go remoteSecretIndexer(indexerCh, resultCh, completeCh, &indexerWg, &resultWg, i)
	}

	go remoteSecretIndexerResultProcessor(&paths, resultCh, completeCh, &resultWg)

	// Wait for all indexers to finish up
	if waitTimeout(&indexerWg, time.Minute*5) {
		log.Fatal("Timeout reached (5m) waiting for remote secret scanner to complete")
	}

	// Wait for all readers to finish up
	if waitTimeout(&resultWg, time.Minute*1) {
		log.Fatal("Timeout reached (5m) waiting for remote secret scanner to complete")
	}

	paths = filterByEnvironment(paths, c.GlobalString("environment"))

	log.Infof("Scanning complete, found %d secrets", len(paths))

	if c.Bool("detailed") {
		printDetailedSecrets(paths)
		return nil
	}

	log.Println()
	for _, secret := range paths {
		log.Infof("%s @ %s: %s", secret.Application, secret.Environment, secret.Path)
	}

	return nil
}

func filterByEnvironment(secrets SecretList, environment string) (result SecretList) {
	if environment == "" {
		return secrets
	}

	for _, s := range secrets {
		if s.Environment == environment {
			result = append(result, s)
		}
	}

	return result
}

func printDetailedSecrets(paths SecretList) {
	secrets, err := readRemoteSecrets(paths)
	if err != nil {
		log.Fatal(err)
	}

	for _, secret := range secrets {
		log.Println()
		log.Infof("%s @ %s: %s", secret.Application, secret.Environment, secret.Path)

		for k, v := range secret.Secret.Data {
			switch vv := v.(type) {
			case string:
				log.Info("  ⇛ ", k, " = ", vv)
			case int:
				log.Println("  ⇛ ", k, " = ", vv)
			default:
				log.Panic("  ⇛ ", k, "is of a type I don't know how to handle")
			}
		}
	}
}

func remoteSecretIndexerResultProcessor(result *SecretList, resultCh chan string, completeCh chan interface{}, wg *sync.WaitGroup) {
	for {
		select {
		case <-completeCh:
			return
		case path := <-resultCh:
			environment, application, err := extraEnvironmentFromPath(path)
			if err != nil {
				environment = "unknown"
				log.Warnf("Could not extract environment from %s", path)
			}

			*result = append(*result, &InternalSecret{path, environment, application, nil})
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
