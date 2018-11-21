package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

type scanner struct {
	config    *Config
	templater *renderer
	path      string
}

func newConfigScanner(directory string, config *Config, templater *renderer) *scanner {
	return &scanner{
		config:    config,
		templater: templater,
		path:      directory,
	}
}

func (s *scanner) scan() error {
	p, err := os.Open(s.path)
	if err != nil {
		return err
	}

	i, err := p.Stat()
	if err != nil {
		return err
	}

	if i.IsDir() {
		return s.scanDirectory(s.path)
	}

	return s.readAndProcess(s.path)
}

// scanDirectory ...
func (s *scanner) scanDirectory(directory string) error {
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
			if err := s.readAndProcess(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, fmt.Errorf("[%s] %s", strings.TrimPrefix(directory, s.path)+"/"+fi.Name(), err))
			}

			continue
		}

		if fi.IsDir() {
			if err := s.scanDirectory(directory + "/" + fi.Name()); err != nil {
				result = multierror.Append(result, err)
			}

			continue
		}

		log.Debugf("Ignoring file %s/%s", directory, fi.Name())
	}

	return result
}

func (s *scanner) readAndProcess(file string) error {
	if strings.HasSuffix(file, ".var.hcl") {
		log.Warnf("Ignoring files with .var.hcl extension")
		return nil
	}

	relativeFile := strings.TrimPrefix(strings.TrimPrefix(file, s.path), "/")

	content, err := s.readFile(file)
	if err != nil {
		return err
	}

	content, err = s.templater.renderContent(content, file, 0)
	if err != nil {
		return err
	}

	if os.Getenv("PRINT_CONTENT") == "1" {
		log.WithField("file", relativeFile).Debug(content)
	}

	list, err := s.config.parseContent(content, relativeFile)
	if err != nil {
		return err
	}

	return s.config.processContent(list, relativeFile)
}

// Read File Content
func (s *scanner) readFile(file string) (string, error) {
	log.Debugf("Parsing file %s", file)

	// read file from disk
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
