package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
		pathName := path.Clean(directory + "/" + fi.Name())

		if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
			linkpath, err := os.Readlink(directory + fi.Name())
			if err != nil {
				return err
			}
			fi, err = os.Lstat(linkpath)
			if err != nil {
				return err
			}
		}

		// regular file
		if fi.Mode().IsRegular() {
			if !s.shouldProcess(pathName) {
				continue
			}

			if err := s.readAndProcess(pathName); err != nil {
				result = multierror.Append(result, fmt.Errorf("[%s] %s", strings.TrimPrefix(directory, pathName), err))
			}

			continue
		}

		// directory
		if fi.IsDir() {
			if err := s.scanDirectory(pathName); err != nil {
				result = multierror.Append(result, err)
			}

			continue
		}

		// something else, ignore it
		log.Debugf("Ignoring path %s/%s", directory, fi.Name())
	}

	return result
}

func (s *scanner) shouldProcess(file string) bool {
	ext := path.Ext(file)

	// we only allow HCL and CTMPL files to be processed
	if ext != ".hcl" && ext != ".ctmpl" {
		log.Debugf("Ignoring file %s (only .hcl or .ctmpl is acceptable file extensions)", file)
		return false
	}

	// files with .var.hcl suffix is considered variable files
	// and should not be processed any further
	if strings.HasSuffix(file, ".var.hcl") {
		log.Debugf("Skipping file %s, is a configuration file", file)
		return false
	}

	// don't process files that was provided as variable files
	// since their syntax is different
	absPath, _ := filepath.Abs(file)
	for _, configFile := range s.templater.readConfigFiles {
		if configFile == absPath {
			log.Debugf("Skipping file %s, is a configuration file", file)
			return false
		}
	}

	return true
}

func (s *scanner) readAndProcess(file string) error {
	content, err := s.readFile(file)
	if err != nil {
		return err
	}

	content, err = s.templater.renderContent(content, file, 0)
	if err != nil {
		return err
	}

	relativeFile := strings.TrimPrefix(strings.TrimPrefix(file, s.path), "/")
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
	log.Debugf("Reading file %s", file)

	// read file from disk
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
