package main

import (
	"fmt"
	"strings"
	"time"

	"sync"

	log "github.com/Sirupsen/logrus"
	api "github.com/hashicorp/vault/api"
	cli "gopkg.in/urfave/cli.v1"
)

func listRemoteSecretsCommand(c *cli.Context) error {
	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	log.Info("start!")

	quitCh := make(chan interface{})
	jobCh := make(chan string, 9999)
	outCh := make(chan string, 9999)

	jobCh <- ""

	var wg sync.WaitGroup

	go remoteSecretIndexer(client, jobCh, outCh, quitCh, &wg, 1)
	go remoteSecretIndexer(client, jobCh, outCh, quitCh, &wg, 2)
	go remoteSecretIndexer(client, jobCh, outCh, quitCh, &wg, 3)
	go remoteSecretIndexer(client, jobCh, outCh, quitCh, &wg, 4)

	time.Sleep(time.Second * 1)
	wg.Wait()
	quitCh <- true

	log.Infof("Number of keys: %d", len(outCh))

	msgCount := len(outCh)
	for i := 0; i < msgCount; i++ {
		log.Printf("path: %s", <-outCh)
	}

	return nil
}

func remoteSecretIndexer(client *api.Client, jobCh chan string, outCh chan string, quitCh chan interface{}, wg *sync.WaitGroup, n int) {
	log.Infof("Starting worker %d", n)

	for {
		select {
		case <-quitCh:
			log.Infof("Stopping worker %d", n)
			return
		case path := <-jobCh:
			wg.Add(1)
			recursiveDecentTree(client, path, jobCh, outCh, n)
			wg.Done()
		}
	}
}

func recursiveDecentTree(client *api.Client, path string, jobCh chan string, outCh chan string, n int) error {
	realPath := fmt.Sprintf("secret/%s", strings.Trim(path, "/"))
	log.Infof("[%d] Process path: %s", n, realPath)

	resp, err := client.Logical().List(realPath)
	if err != nil {
		return err
	}

	if resp.Data == nil {
		return fmt.Errorf("Response contains no data")
	}

	rawKeys, ok := resp.Data["keys"]
	if !ok {
		return fmt.Errorf("Could not find any keys in the response, server issues?")
	}

	keys := secretsToString(rawKeys)

	if len(keys) == 0 {
		return fmt.Errorf("No keys found in the vault")
	}

	for _, v := range keys {
		if v[len(v)-1:] == "/" {
			jobCh <- fmt.Sprintf("%s/%s", path, v[0:len(v)-1])
			continue
		}

		outCh <- fmt.Sprintf("%s/%s", path, v)
	}

	return nil
}

func secretsToString(in interface{}) (out []string) {
	t := in.([]interface{})
	for _, v := range t {
		out = append(out, v.(string))
	}

	return out
}
