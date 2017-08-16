package vault

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
    log "github.com/Sirupsen/logrus"

	cli "gopkg.in/urfave/cli.v1"
)

// EditProfile ...
func EditProfile(c *cli.Context) error {
	filePath := getProfileFile()

	file, err := ioutil.TempFile("", "vault")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	backup := false

	// if the file exist, put in the original encrypted content
	if _, err := os.Stat(filePath); err == nil {
		backup = true

		b, err := getProfileConfig()
		if err != nil {
			return err
		}
		file.Write(b)
	}

	// find the editor to use from env, default to pico/nano
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "pico"
	}

	// custom flags to weird editors
	flags := make([]string, 0)
	switch editor {
	case "code":
		flags = append(flags, "-w")
		flags = append(flags, "-n")
	}

	// append the filename
	flags = append(flags, file.Name())

	// edit the file
	editCmd := exec.Command(editor, flags...)
	editErr := editCmd.Run()
	if editErr != nil {
		log.Info("editErr %s\n", editErr)
		log.Info(editErr)
		return editErr
	}

	log.Info("finish edit%s\n", editCmd)

	// backup the old file
	if backup {
		copyFileContents(getProfileFile(), getProfileFile()+".old")
	}

	// encrypt the file
	encryptCmd := exec.Command("keybase", "pgp", "encrypt", "--infile", file.Name(), "--outfile", getProfileFile())
	encryptErr := encryptCmd.Run()
	if encryptErr != nil {
		return encryptErr
	}

	return nil
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
