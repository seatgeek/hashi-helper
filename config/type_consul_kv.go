package config

import "github.com/hashicorp/consul/api"
import "fmt"

// ConsulKV ...
type ConsulKV struct {
	Application *Application
	Environment *Environment
	Key         string
	Value       []byte
}

// ToConsulKV ...
func (c *ConsulKV) ToConsulKV() *api.KVPair {
	return &api.KVPair{
		Key:   c.toPath(),
		Value: c.Value,
	}
}

func (c *ConsulKV) toPath() string {
	if c.Application != nil {
		return fmt.Sprintf("%v/%v", c.Application.Name, c.Key)
	}
	return c.Key
}

// ConsulKV struct
//
type ConsulKVs []*ConsulKV

// Add ...
func (cs *ConsulKVs) Add(kv *ConsulKV) {
	*cs = append(*cs, kv)
}
