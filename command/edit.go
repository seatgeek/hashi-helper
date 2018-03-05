package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	cli "gopkg.in/urfave/cli.v1"
	log "github.com/Sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	vault "github.com/seatgeek/hashi-helper/command/vault"
)

type preferences map[string]preference
type preference struct {
	KeybaseTeam  string `yaml:"keybase-team-name"`
	Editor string `yaml:"editor"`
}

// EditEncryptedFile ...
func EditEncryptedFile(c *cli.Context) error {

	// identify the target file
	if !c.Args().Present() {
		return fmt.Errorf("Please provide a path and file name.")
	}
	filePath := c.Args().First()
	if filePath == "" {
		return fmt.Errorf("Empty path and file name.")
	}

	// get the preferences, define the editor and keybase team
	var preferences preferences
	configFile, err := vault.GetProfileConfig()
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(configFile, &preferences); err != nil {
		return err
	}
	preference, ok := preferences["encrypted-file-config"]
	if !ok {
		return fmt.Errorf("No encrypted-file-config block found in config file.\n")
	}
	// editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = preference.Editor
	}
	if editor == "" {
		editor = "pico"
	}
	// keybase team
	var keybaseTeam string = preference.KeybaseTeam
	if keybaseTeam == "" && NoKeybaseTeam == false {
		log.Warnf("No Keybase Team found in config (and --no-keybase-team is unset). Not using Keybase Team.\n")
	}
	if NoKeybaseTeam == true {
		keybaseTeam = ""
	}

	// create the temporary file
	tempFile, err := ioutil.TempFile("", "encrypted-hcl")
	if err != nil {
		return err
	}

	// if the target file exist, decrypt the contents and copy into the temp file
	if _, err := os.Stat(filePath); err == nil {

		cmd := exec.Command("keybase", "pgp", "decrypt", "--infile", filePath)

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		log.Infof("Starting keybase decrypt of %s\n", filePath)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Failed to run keybase gpg decrypt: %s - %s",
				                    err, stderr.String())
		}

		tempFile.Write(stdout.Bytes())
	}

	// edit the temp file
	editCmd := exec.Command(editor, tempFile.Name())
	editCmd.Stdin  = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr
	editErr := editCmd.Run()
	if editErr != nil {
		return editErr
	}

	// editing complete. replace the original file with the temp file and re-encrypt
	// optionally with keybase team members
	encryptCmd := exec.Command("keybase", "pgp", "encrypt",
		                         "--i", tempFile.Name(),
	                           "--o", filePath)
	log.Infof("Re-encrypting updated %s.\n", filePath)
	if NoKeybaseTeam == false { // I know, a double negative. Sue me.
		cmd := "keybase team list-members " + keybaseTeam + " | awk '{print $3}' | tr '\\n' ' ' | xargs | tr -d '\\n'"
    memberListCmd := exec.Command("bash", "-c", cmd)
		var stdout bytes.Buffer
		memberListCmd.Stdout = &stdout

		var stderr bytes.Buffer
		memberListCmd.Stderr = &stderr

		memberListErr := memberListCmd.Run()
		if memberListErr != nil {
			return memberListErr
		}

		memberList := stdout.String()
		cmd = "keybase pgp encrypt " + memberList + " --i " + tempFile.Name() + " --o " + filePath
		encryptCmd = exec.Command("bash", "-c", cmd)
	  log.Infof("Using Keybase team \"%s\" with members \"%s\".\n", keybaseTeam, memberList)
	}
	encryptErr := encryptCmd.Run()
	if encryptErr != nil {
		return encryptErr
	}

	// defered action to remove temporary file post-edit and post-re-encryption
	defer os.Remove(tempFile.Name())

	return nil
}

