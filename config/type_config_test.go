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

			got, err := c.parseContent(tt.content)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err2 := c.processContent(got)
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
			template:     "hello world",
			wantTemplate: "hello world",
		},
		{
			name:     "test service func missing consul_domain",
			template: `[[ service "derp" ]]`,
			wantErr:  errors.New("Missing interpolation key 'consul_domain'"),
		},
		{
			name:     "test template func: service",
			template: `[[ service "vault" ]]`,
			templateVariables: map[string]interface{}{
				"consul_domain": "consul",
			},
			wantTemplate: "vault.service.consul",
		},
		{
			name:     "test template func:  service_with_tag",
			template: `[[ service_with_tag "vault" "active" ]]`,
			templateVariables: map[string]interface{}{
				"consul_domain": "consul",
			},
			wantTemplate: "active.vault.service.consul",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				templateVariables: tt.templateVariables,
			}

			got, err := c.renderContent(tt.template)
			if tt.wantErr != nil {
				require.True(t, strings.Contains(err.Error(), tt.wantErr.Error()))
				require.Equal(t, "", tt.wantTemplate, "you should not expect a template during error tests")
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantTemplate, got)
		})
	}
}
