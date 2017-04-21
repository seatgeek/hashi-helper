package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/vault/helper/parseutil"
)

//
// COPIED FROM https://github.com/hashicorp/vault/blob/master/vault/policy.go
// EXCLUDING "Parse" METHOD WHICH IS CALLED parsePolicyStanza IN OUR parser.go
//

const (
	DenyCapability   = "deny"
	CreateCapability = "create"
	ReadCapability   = "read"
	UpdateCapability = "update"
	DeleteCapability = "delete"
	ListCapability   = "list"
	SudoCapability   = "sudo"
	RootCapability   = "root"

	// Backwards compatibility
	OldDenyPathPolicy  = "deny"
	OldReadPathPolicy  = "read"
	OldWritePathPolicy = "write"
	OldSudoPathPolicy  = "sudo"
)

const (
	DenyCapabilityInt uint32 = 1 << iota
	CreateCapabilityInt
	ReadCapabilityInt
	UpdateCapabilityInt
	DeleteCapabilityInt
	ListCapabilityInt
	SudoCapabilityInt
)

var (
	cap2Int = map[string]uint32{
		DenyCapability:   DenyCapabilityInt,
		CreateCapability: CreateCapabilityInt,
		ReadCapability:   ReadCapabilityInt,
		UpdateCapability: UpdateCapabilityInt,
		DeleteCapability: DeleteCapabilityInt,
		ListCapability:   ListCapabilityInt,
		SudoCapability:   SudoCapabilityInt,
	}
)

// PathCapabilities represents a policy for a path in the namespace.
type PathCapabilities struct {
	Prefix       string
	Policy       string
	Permissions  *Permissions
	Glob         bool
	Capabilities []string

	// These keys are used at the top level to make the HCL nicer; we store in
	// the Permissions object though
	MinWrappingTTLHCL    interface{}              `hcl:"min_wrapping_ttl"`
	MaxWrappingTTLHCL    interface{}              `hcl:"max_wrapping_ttl"`
	AllowedParametersHCL map[string][]interface{} `hcl:"allowed_parameters"`
	DeniedParametersHCL  map[string][]interface{} `hcl:"denied_parameters"`
}

type Permissions struct {
	CapabilitiesBitmap uint32
	MinWrappingTTL     time.Duration
	MaxWrappingTTL     time.Duration
	AllowedParameters  map[string][]interface{}
	DeniedParameters   map[string][]interface{}
}

func parsePaths(result *Policy, list *ast.ObjectList) error {
	paths := make([]*PathCapabilities, 0, len(list.Items))
	for _, item := range list.Items {
		key := "path"
		if len(item.Keys) > 0 {
			key = item.Keys[0].Token.Value().(string)
		}
		valid := []string{
			"policy",
			"capabilities",
			"allowed_parameters",
			"denied_parameters",
			"min_wrapping_ttl",
			"max_wrapping_ttl",
		}
		if err := checkHCLKeys(item.Val, valid); err != nil {
			return multierror.Prefix(err, fmt.Sprintf("path %q:", key))
		}

		var pc PathCapabilities

		// allocate memory so that DecodeObject can initialize the Permissions struct
		pc.Permissions = new(Permissions)

		pc.Prefix = key
		if err := hcl.DecodeObject(&pc, item.Val); err != nil {
			return multierror.Prefix(err, fmt.Sprintf("path %q:", key))
		}

		// Strip a leading '/' as paths in Vault start after the / in the API path
		if len(pc.Prefix) > 0 && pc.Prefix[0] == '/' {
			pc.Prefix = pc.Prefix[1:]
		}

		// Strip the glob character if found
		if strings.HasSuffix(pc.Prefix, "*") {
			pc.Prefix = strings.TrimSuffix(pc.Prefix, "*")
			pc.Glob = true
		}

		// Map old-style policies into capabilities
		if len(pc.Policy) > 0 {
			switch pc.Policy {
			case OldDenyPathPolicy:
				pc.Capabilities = []string{DenyCapability}
			case OldReadPathPolicy:
				pc.Capabilities = append(pc.Capabilities, []string{ReadCapability, ListCapability}...)
			case OldWritePathPolicy:
				pc.Capabilities = append(pc.Capabilities, []string{CreateCapability, ReadCapability, UpdateCapability, DeleteCapability, ListCapability}...)
			case OldSudoPathPolicy:
				pc.Capabilities = append(pc.Capabilities, []string{CreateCapability, ReadCapability, UpdateCapability, DeleteCapability, ListCapability, SudoCapability}...)
			default:
				return fmt.Errorf("path %q: invalid policy '%s'", key, pc.Policy)
			}
		}

		// Initialize the map
		pc.Permissions.CapabilitiesBitmap = 0
		for _, cap := range pc.Capabilities {
			switch cap {
			// If it's deny, don't include any other capability
			case DenyCapability:
				pc.Capabilities = []string{DenyCapability}
				pc.Permissions.CapabilitiesBitmap = DenyCapabilityInt
				goto PathFinished
			case CreateCapability, ReadCapability, UpdateCapability, DeleteCapability, ListCapability, SudoCapability:
				pc.Permissions.CapabilitiesBitmap |= cap2Int[cap]
			default:
				return fmt.Errorf("path %q: invalid capability '%s'", key, cap)
			}
		}

		if pc.AllowedParametersHCL != nil {
			pc.Permissions.AllowedParameters = make(map[string][]interface{}, len(pc.AllowedParametersHCL))
			for key, val := range pc.AllowedParametersHCL {
				pc.Permissions.AllowedParameters[strings.ToLower(key)] = val
			}
		}
		if pc.DeniedParametersHCL != nil {
			pc.Permissions.DeniedParameters = make(map[string][]interface{}, len(pc.DeniedParametersHCL))

			for key, val := range pc.DeniedParametersHCL {
				pc.Permissions.DeniedParameters[strings.ToLower(key)] = val
			}
		}
		if pc.MinWrappingTTLHCL != nil {
			dur, err := parseutil.ParseDurationSecond(pc.MinWrappingTTLHCL)
			if err != nil {
				return errwrap.Wrapf("error parsing min_wrapping_ttl: {{err}}", err)
			}
			pc.Permissions.MinWrappingTTL = dur
		}
		if pc.MaxWrappingTTLHCL != nil {
			dur, err := parseutil.ParseDurationSecond(pc.MaxWrappingTTLHCL)
			if err != nil {
				return errwrap.Wrapf("error parsing max_wrapping_ttl: {{err}}", err)
			}
			pc.Permissions.MaxWrappingTTL = dur
		}
		if pc.Permissions.MinWrappingTTL != 0 &&
			pc.Permissions.MaxWrappingTTL != 0 &&
			pc.Permissions.MaxWrappingTTL < pc.Permissions.MinWrappingTTL {
			return errors.New("max_wrapping_ttl cannot be less than min_wrapping_ttl")
		}

	PathFinished:
		paths = append(paths, &pc)
	}

	result.Paths = paths
	return nil
}
