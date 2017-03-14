package main

import (
	"log"

	c "github.com/seatgeek/vault-restore/config"
)

func main() {
	config, err := c.NewConfigFromDirectory("./conf.d")
	if err != nil {
		log.Fatalf(err.Error())
	}

	for envName, apps := range config {
		log.Printf("Environment: %s (apps: %d)", envName, len(apps))

		for appName, app := range apps {
			log.Printf("  app: %s", appName)

			for secretKey, secretValues := range app.Secrets {
				log.Printf("    secret: %s", secretKey)

				for k, v := range secretValues {
					log.Printf("      %s = %s", k, v)
				}
			}
		}

		log.Println()
	}

	// spew.Dump(config)
	// spew.Dump(err)
}
