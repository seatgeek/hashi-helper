package main

import (
	log "github.com/Sirupsen/logrus"
	cfg "github.com/seatgeek/vault-restore/config"
)

func main() {
	config, err := cfg.NewConfigFromDirectory("./conf.d")
	if err != nil {
		log.Fatalf(err.Error())
	}

	for envName, apps := range config {
		envLogger := log.WithFields(log.Fields{"env": envName})

		for appName, app := range apps {
			appLogger := envLogger.WithFields(log.Fields{"app": appName})

			for secretKey, secretValues := range app.Secrets {
				secretLogger := appLogger.WithFields(log.Fields{"secret": secretKey})

				for k, v := range secretValues {
					secretLogger.Printf("%s = %s", k, v)
				}
			}
		}

		log.Println()
	}

	// spew.Dump(config)
	// spew.Dump(err)
}
