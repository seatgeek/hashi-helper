package config

import (
	"fmt"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// Secret ...
type Audit struct {
	Description string `hcl:"description"`
	Environment *Environment
	Key         string
	Local       bool                   `hcl:"local"`
	Options     map[string]interface{} `hcl:"options"`
	Path        string                 `hcl:"path"`
	Type        string                 `hcl:"type"`
}

// Equal ...
func (s *Audit) Equal(o *Audit) bool {
	return s.Path == o.Path && s.Key == o.Key
}

func (s *Audit) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"description": s.Description,
		"options":     s.Options,
		"type":        s.Type,
		"local":       s.Local,
	}
}

// VaultAudits struct
//
// environment -> application
type VaultAudits []*Audit

// Add ...
func (e *VaultAudits) Add(audit *Audit) bool {
	if !e.Exists(audit) {
		*e = append(*e, audit)
		return true
	}

	return false
}

// Exists ...
func (e *VaultAudits) Exists(audit *Audit) bool {
	for _, existing := range *e {
		if audit.Equal(existing) {
			return true
		}
	}

	return false
}

// Get ...
func (e *VaultAudits) Get(audit *Audit) *Audit {
	for _, existing := range *e {
		if audit.Equal(existing) {
			return existing
		}
	}

	return nil
}

// GetOrSet ...
func (e *VaultAudits) GetOrSet(audit *Audit) *Audit {
	existing := e.Get(audit)
	if existing != nil {
		return existing
	}

	e.Add(audit)
	return audit
}

func (e *VaultAudits) List() []string {
	res := []string{}

	for _, sec := range *e {
		res = append(res, sec.Path)
	}

	return res
}

// parseVaultAuditStanza
// parse out `environment -> audit
func (c *Config) parseVaultAuditStanza(list *ast.ObjectList, env *Environment) error {
	if len(list.Items) == 0 {
		return nil
	}

	c.logger = c.logger.WithField("stanza", "audit")
	c.logger.Debugf("Found %d audit{}", len(list.Items))
	for _, auditData := range list.Items {
		if len(auditData.Keys) != 1 {
			return fmt.Errorf("Missing audit name in line %+v", auditData.Keys[0].Pos())
		}

		var audit Audit
		if err := hcl.DecodeObject(&audit, auditData); err != nil {
			return err
		}

		auditName := auditData.Keys[0].Token.Value().(string)

		audit.Key = auditName
		audit.Path = auditName
		audit.Environment = env

		if c.VaultAudits.Add(&audit) == false {
			c.logger.Warnf("Ignored duplicate audit '%s' -> '%s' in line %s", audit.Environment.Name, audit.Key, auditData.Keys[0].Token.Pos)
		}
	}

	return nil
}
