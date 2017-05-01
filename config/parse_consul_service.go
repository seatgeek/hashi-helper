package config

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/hcl/hcl/ast"
)

func (c *Config) processConsulServices(list *ast.ObjectList, environment *Environment) error {
	if len(list.Items) == 0 {
		return nil
	}

	for _, serviceAST := range list.Items {
		x := serviceAST.Val.(*ast.ObjectType).List

		valid := []string{"address", "node", "port", "tags"}
		if err := checkHCLKeys(x, valid); err != nil {
			return err
		}

		if len(serviceAST.Keys) != 1 {
			return fmt.Errorf("Missing service name in line %+v", serviceAST.Keys[0].Pos())
		}

		serviceName := serviceAST.Keys[0].Token.Value().(string)

		address, err := getKeyString("address", x)
		if err != nil {
			return err
		}

		node, err := getKeyString("node", x)
		if err != nil {
			return err
		}

		port, err := getKeyInt("port", x)
		if err != nil {
			return err
		}

		tags, err := getKeyStringList("tags", x)
		if err != nil {
			if strings.Contains(err.Error(), "missing tags") {
				tags = make([]string, 0)
			} else {
				return err
			}
		}

		serviceID, err := getKeyString("id", x)
		if err != nil {
			if strings.Contains(err.Error(), "missing id") {
				serviceID = serviceName
			} else {
				return err
			}
		}

		service := &ConsulService{
			Node:    node,
			Address: address,
			Service: &api.AgentService{
				Address: address,
				ID:      serviceID,
				Port:    port,
				Service: serviceName,
				Tags:    tags,
			},
			Check: &api.AgentCheck{
				CheckID:     fmt.Sprintf("service:%s", serviceName),
				Name:        serviceName,
				Node:        node,
				Notes:       "created by hashi-helper",
				ServiceName: serviceName,
				ServiceID:   serviceID,
				Status:      "passing",
			},
		}

		c.ConsulServices.Add(service)
	}

	return nil
}

func getKeyString(key string, x *ast.ObjectList) (string, error) {
	list := x.Filter(key)
	if len(list.Items) == 0 {
		return "", fmt.Errorf("missing %s", key)
	}

	if len(list.Items) > 1 {
		return "", fmt.Errorf("More than one match for %s", key)
	}

	value := list.Items[0].Val.(*ast.LiteralType).Token.Value().(string)

	return value, nil
}

func getKeyStringList(key string, x *ast.ObjectList) ([]string, error) {
	list := x.Filter(key)
	if len(list.Items) != 1 {
		return nil, fmt.Errorf("missing %s", key)
	}

	z := list.Items[0].Val.(*ast.ListType)

	res := make([]string, 0)
	for _, i := range z.List {
		val := i.(*ast.LiteralType).Token.Value().(string)
		res = append(res, val)
	}

	return res, nil
}

func getKeyInt(key string, x *ast.ObjectList) (int, error) {
	list := x.Filter(key)
	if len(list.Items) == 0 {
		return 0, fmt.Errorf("missing %s", key)
	}

	if len(list.Items) > 1 {
		return 0, fmt.Errorf("More than one match for %s", key)
	}

	value := int(list.Items[0].Val.(*ast.LiteralType).Token.Value().(int64))

	return value, nil
}
