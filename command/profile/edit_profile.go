package profile

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"gopkg.in/urfave/cli.v1"
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
	} else {
		b := []byte(`---
# Sample config (yaml)
#
# profile_name_1:
#   vault:
#     server: http://active.vault.service.consul:8200
#     auth:
#         token: <your vault token>
#         unseal_token: <your unseal token>
#   consul:
#     server: http://consul.service.consul:8500
#     auth:
#         token: <your consul token>
#   nomad:
#     server: http://nomad.service.consul:4646
#     auth:
#         token: <your nomad token>
#
# profile_name_2:
#   vault:
#     server: http://active.vault.service.consul:8200
#     auth:
#         method: github
#         github_token: <your github token>
#   consul:
#     server: http://consul.service.consul:8500
#     auth:
#       method: vault
#       creds_path: consul/creds/administrator
#   nomad:
#     server: http://nomad.service.consul:4646
#     auth:
#       method: vault
#       creds_path: nomad/creds/administrator

`)
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
	// More Editors should be added
	}

	// append the filename
	flags = append(flags, file.Name())

	// edit the file
	editCmd := exec.Command(editor, flags...)
	editErr := editCmd.Run()
	if editErr != nil {
		return editErr
	}

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
