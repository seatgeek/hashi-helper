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
func remoteSecretIndexer(jobCh chan string, outCh chan string, quitCh chan interface{}, wg *sync.WaitGroup, resultWg *sync.WaitGroup, n int) {
	log.WithField("method", "remoteSecretIndexer").Debugf("Starting worker %d", n)

	// Create a new Vault API client for this go-routine
	client, err := api.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-quitCh:
			log.WithField("method", "remoteSecretIndexer").Debugf("Stopping worker %d", n)
			return
		case path := <-jobCh:
			realPath := fmt.Sprintf("secret/%s", strings.Trim(path, "/"))
			log.WithField("method", "remoteScanPathRecursive").Debugf("[%d] Scanning path: %s", n, realPath)

			resp, err := client.Logical().List(realPath)
			if err != nil {
				log.Fatal(err)
			}

			if resp.Data == nil {
				log.Fatal("Response contains no data")
			}

			rawKeys, ok := resp.Data["keys"]
			if !ok {
				log.Fatal("Could not find any keys in the response, server issues?")
			}

			keys := secretsToString(rawKeys)

			if len(keys) == 0 {
				log.Fatal("No keys found in the vault")
			}

			for _, v := range keys {
				if v[len(v)-1:] == "/" {
					wg.Add(1)
					jobCh <- fmt.Sprintf("%s/%s", path, v[0:len(v)-1])
					continue
				}

				resultWg.Add(1)
				outCh <- fmt.Sprintf("%s/%s", realPath, v)
			}

			wg.Done()
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
