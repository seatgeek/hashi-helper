package config

import "github.com/hashicorp/consul/api"

// ConsulService ...
type ConsulService api.CatalogRegistration

// ToConsulService ...
func (c *ConsulService) ToConsulService() *api.CatalogRegistration {
	return &api.CatalogRegistration{
		ID:      c.ID,
		Node:    c.Node,
		Address: c.Address,
		Service: c.Service,
		Check:   c.Check,
	}
}

// ConsulServices struct
//
type ConsulServices []*ConsulService

// Add ...
func (cs *ConsulServices) Add(service *ConsulService) {
	*cs = append(*cs, service)
}
