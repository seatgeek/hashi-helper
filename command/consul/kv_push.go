package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/seatgeek/hashi-helper/config"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
)

// KVPush ...
func KVPush(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	return KVPushWithConfig(c, config)
}

// KVPushWithConfig ...
func KVPushWithConfig(c *cli.Context, config *config.Config) error {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	kvService := client.KV()

	for _, kv := range config.ConsulKVs {
		log.Infof("Saving consul KV %s", kv.Key)

		consulKV := kv.ToConsulKV()

		meta, err := kvService.Put(consulKV, &api.WriteOptions{})
		if err != nil {
			return err
		}

		log.Infof("  Saved KV in %s", meta.RequestTime.String())
	}

	return nil
}
