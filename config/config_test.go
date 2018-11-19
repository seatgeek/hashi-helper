package config

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ParseContent(t *testing.T) {
	tests := []struct {
		name             string
		env              string
		content          string
		seenEnvironments []string
		seenApplications []string
		seenSecrets      []string
		wantErr          bool
	}{
		// wildcard and named environment mixed, should expose the seen environment
		// as the "test" since "*" matches that
		{
			name: "parse simple",
			env:  "test",
			content: `
environment "*" {
	application "seatgeek" {
		secret "very-secret" {
			value = "hello world"
		}
	}
}`,
			seenEnvironments: []string{"test"},
			seenApplications: []string{"seatgeek"},
			seenSecrets:      []string{"very-secret"},

			wantErr: false,
		},
		//
		{
			name: "parse multi with match",
			env:  "prod",
			content: `
environment "prod" "stag" {
	application "seatgeek" {
		secret "very-secret" {
			value = "hello world"
		}
	}
}`,
			seenEnvironments: []string{"prod"},
			seenApplications: []string{"seatgeek"},
			seenSecrets:      []string{"very-secret"},

			wantErr: false,
		},
		{
			name: "parse multi with _no_ match",
			env:  "perf",
			content: `
environment "prod" "stag" {
	application "seatgeek" {
		secret "very-secret" {
			value = "hello world"
		}
	}
}`,
			seenEnvironments: []string{},
			seenApplications: []string{},
			seenSecrets:      []string{},

			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}

			TargetEnvironment = tt.env

			got, err := c.parseContent(tt.content, "test.hcl")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err2 := c.processContent(got, "test.hcl")
			if tt.wantErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
			}

			require.Equal(t, tt.seenEnvironments, c.Environments.List())
			require.Equal(t, tt.seenApplications, c.Applications.List())
			require.Equal(t, tt.seenSecrets, c.VaultSecrets.List())
		})
	}
}

func TestConfig_renderContent(t *testing.T) {
	tests := []struct {
		name              string
		template          string
		templateVariables map[string]interface{}
		wantTemplate      string
		wantErr           error
	}{
		{
			name:         "no templating, passthrough",
			template:     `hello="world"`,
			wantTemplate: `hello = "world"`,
		},
		{
			name:     "test service func missing consul_domain",
			template: `[[ service "derp" ]]`,
			wantErr:  errors.New("Missing interpolation key 'consul_domain'"),
		},
		{
			name:     "test template func: service",
			template: `service="[[ service "vault" ]]"`,
			templateVariables: map[string]interface{}{
				"consul_domain": "consul",
			},
			wantTemplate: `service = "vault.service.consul"`,
		},
		{
			name:     "test template func: service_with_tag",
			template: `service="[[ service_with_tag "vault" "active" ]]"`,
			templateVariables: map[string]interface{}{
				"consul_domain": "consul",
			},
			wantTemplate: `service = "active.vault.service.consul"`,
		},
		{
			name:     "test template func: grant_credentials",
			template: `[[ grant_credentials "my-db" "full" ]]`,
			wantTemplate: `
path "my-db/creds/full" {
  capabilities = ["read"]
}`,
		},
		{
			name:     "test template func: github_assign_team_policy",
			template: `[[ github_assign_team_policy "my-team" "my-policy" ]]`,
			wantTemplate: `
secret "/auth/github/map/teams/my-team" {
  value = "my-policy"
}`,
		},
		{
			name:     "test template func: ldap_assign_group_policy",
			template: `[[ ldap_assign_group_policy "my-group" "my-policy" ]]`,
			wantTemplate: `
secret "/auth/ldap/groups/my-group" {
  value = "my-policy"
}`,
		},
		{
			name:     "test template func: grant_credentials_policy",
			template: `[[ grant_credentials_policy "my-db" "full" ]]`,
			wantTemplate: `
policy "my-db-full" {
  path "my-db/creds/full" {
    capabilities = ["read"]
  }
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templater := &templater{
				templateVariables: tt.templateVariables,
			}

			got, err := templater.renderContent(tt.template, "test", 0)
			if tt.wantErr != nil {
				require.True(t, strings.Contains(err.Error(), tt.wantErr.Error()))
				require.Equal(t, "", tt.wantTemplate, "you should not expect a template during error tests")
				return
			}

			require.NoError(t, err)
			require.Equal(t, strings.TrimSpace(tt.wantTemplate), got)
		})
	}
}
