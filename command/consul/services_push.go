package consul

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"github.com/seatgeek/hashi-helper/config"
	cli "gopkg.in/urfave/cli.v1"
)

// ServicesPush ...
func ServicesPush(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	return ServicesPushWithConfig(c, config)
}

// ServicesPushWithConfig ...
func ServicesPushWithConfig(c *cli.Context, config *config.Config) error {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	catalog := client.Catalog()

	for _, service := range config.ConsulServices {
		log.Infof("Saving consul service %s/%s", service.Node, service.Service.Service)

		consulService := service.ToConsulService()

		meta, err := catalog.Register(consulService, &api.WriteOptions{})
		if err != nil {
			return err
		}

		log.Infof("  Saved service in %s", meta.RequestTime.String())
	}

	return nil
}
