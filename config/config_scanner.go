package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	multierror "github.com/hashicorp/go-multierror"
)

type configScanner struct {
	config    *Config
	templater *templater
	path      string
}

func newConfigScanner(directory string, config *Config, templater *templater) *configScanner {
	return &configScanner{
		config:    config,
		templater: templater,
		path:      directory,
	}
}

func (cs *configScanner) scan() error {
	p, err := os.Open(cs.path)
	if err != nil {
		return err
	}

	i, err := p.Stat()
	if err != nil {
		return err
	}

	if i.IsDir() {
		return cs.scanDirectory(cs.path)
	}

	return cs.readAndProcess(cs.path)
}

// scanDirectory ...
func (cs *configScanner) scanDirectory(directory string) error {
	log.Debugf("Scanning directory %s", directory)

	d, err := os.Open(directory)
	if err != nil {
		return err
	}
	d.Close()

	fi, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	var result error
	for _, fi := range fi {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".hcl") {
			if err := cs.readAndProcess(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, fmt.Errorf("[%s] %s", strings.TrimPrefix(directory, cs.path)+"/"+fi.Name(), err))
			}

			continue
		}

		if fi.IsDir() {
			if err := cs.scanDirectory(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, err)
			}

			continue
		}

		log.Debugf("Ignoring file %s/%s", directory, fi.Name())
	}

	return result
}

func (cs *configScanner) readAndProcess(file string) error {
	if strings.HasSuffix(file, ".var.hcl") {
		log.Warnf("Ignoring files with .var.hcl extension")
		return nil
	}

	relativeFile := strings.TrimPrefix(strings.TrimPrefix(file, cs.path), "/")

	content, err := cs.readFile(file)
	if err != nil {
		return err
	}

	content, err = cs.templater.renderContent(content, file, 0)
	if err != nil {
		return err
	}

	if os.Getenv("PRINT_CONTENT") == "1" {
		log.WithField("file", relativeFile).Debug(content)
	}

	list, err := cs.config.parseContent(content, relativeFile)
	if err != nil {
		return err
	}

	return cs.config.processContent(list, relativeFile)
}

// Read File Content
func (cs *configScanner) readFile(file string) (string, error) {
	log.Debugf("Parsing file %s", file)

	// read file from disk
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
