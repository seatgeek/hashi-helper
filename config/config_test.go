package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
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
		seenServices     ConsulServices
		parserErr        error
		processErr       error
	}{
		// wildcard and named environment mixed, should expose the seen environment
		// as the "test" since "*" matches that
		{
			name:             "parse simple",
			env:              "test",
			seenEnvironments: []string{"test"},
			seenApplications: []string{"seatgeek"},
			seenSecrets:      []string{"very-secret"},
			content: `
environment "*" {
	application "seatgeek" {
		secret "very-secret" {
			value = "hello world"
		}
	}
}`,
		},
		{
			name:             "parse multi with match",
			env:              "prod",
			seenEnvironments: []string{"prod"},
			seenApplications: []string{"seatgeek"},
			seenSecrets:      []string{"very-secret"},
			content: `
environment "prod" "stag" {
	application "seatgeek" {
		secret "very-secret" {
			value = "hello world"
		}
	}
}`,
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
		},
		{
			name:             "parse service{} with meta",
			env:              "perf",
			seenEnvironments: []string{"perf"},
			seenServices: ConsulServices{
				{
					Address: "127.0.0.1",
					Node:    "test",
					Service: &api.AgentService{
						ID:      "test",
						Service: "test",
						Tags:    []string{},
						Port:    1337,
						Address: "127.0.0.1",
						Meta: map[string]string{
							"meta_key_1": "meta_value_1",
							"meta_key_2": "meta_value_2",
						},
					},
					Check: &api.AgentCheck{
						Node:        "test",
						CheckID:     "service:test",
						Name:        "test",
						Status:      "passing",
						Notes:       "created by hashi-helper",
						ServiceID:   "test",
						ServiceName: "test",
					},
				},
			},
			content: `
environment "*" {
	service "test" {
		address = "127.0.0.1"
		node    = "test"
		port    = 1337

		meta {
			meta_key_1 = "meta_value_1"
			meta_key_2 = "meta_value_2"
		}
	}
}`,
		},
		{
			name:             "parse service{} with empty meta",
			env:              "perf",
			seenEnvironments: []string{"perf"},
			seenServices: ConsulServices{
				{
					Address: "127.0.0.1",
					Node:    "test",
					Service: &api.AgentService{
						ID:      "test",
						Service: "test",
						Tags:    []string{},
						Port:    1337,
						Address: "127.0.0.1",
						Meta:    map[string]string{},
					},
					Check: &api.AgentCheck{
						Node:        "test",
						CheckID:     "service:test",
						Name:        "test",
						Status:      "passing",
						Notes:       "created by hashi-helper",
						ServiceID:   "test",
						ServiceName: "test",
					},
				},
			},
			content: `
environment "*" {
	service "test" {
		address = "127.0.0.1"
		node    = "test"
		port    = 1337

		meta {}
	}
}`,
		},
		{
			name:             "process service{} with 2 meta should fail",
			env:              "perf",
			seenEnvironments: []string{"perf"},
			processErr:       fmt.Errorf("You can only specify meta{} once at -"),
			content: `
environment "*" {
	service "test" {
		address = "127.0.0.1"
		node    = "test"
		port    = 1337

		meta {
			meta_key = "meta_value"
		}

		meta {
			meta_key_2 = "meta_value_2"
		}
	}
}`,
		},
		{
			name:             "parse service{} with no meta",
			env:              "perf",
			seenEnvironments: []string{"perf"},
			seenServices: ConsulServices{
				{
					Address: "127.0.0.1",
					Node:    "test",
					Service: &api.AgentService{
						ID:      "test",
						Service: "test",
						Tags:    []string{},
						Port:    1337,
						Address: "127.0.0.1",
					},
					Check: &api.AgentCheck{
						Node:        "test",
						CheckID:     "service:test",
						Name:        "test",
						Status:      "passing",
						Notes:       "created by hashi-helper",
						ServiceID:   "test",
						ServiceName: "test",
					},
				},
			},
			content: `
environment "*" {
	service "test" {
		address = "127.0.0.1"
		node    = "test"
		port    = 1337
	}
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				targetEnvironment: tt.env,
			}

			got, err := c.parseContent(tt.content, "test.hcl")
			if tt.parserErr != nil {
				require.EqualError(t, err, tt.parserErr.Error())
			} else {
				require.NoError(t, err)
			}

			err2 := c.processContent(got, "test.hcl")
			if tt.processErr != nil {
				require.EqualError(t, err2, tt.processErr.Error())
			} else {
				require.NoError(t, err2)
			}

			require.Equal(t, tt.seenEnvironments, c.Environments.list())
			require.Equal(t, tt.seenApplications, c.Applications.list())
			require.Equal(t, tt.seenSecrets, c.VaultSecrets.List())
			require.Equal(t, tt.seenServices, c.ConsulServices.List())
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
			name:         "test service func missing consul_domain",
			template:     `service = "[[ service "derp" ]]"`,
			wantTemplate: `service = "derp.service.consul"`,
		},
		{
			name:     "test template func: service",
			template: `service="[[ service "vault" ]]"`,
			templateVariables: map[string]interface{}{
				"consul_domain": "test.consul",
			},
			wantTemplate: `service = "vault.service.test.consul"`,
		},
		{
			name:     "test template func: serviceWithTag",
			template: `service="[[ serviceWithTag "vault" "active" ]]"`,
			templateVariables: map[string]interface{}{
				"consul_domain": "test.consul",
			},
			wantTemplate: `service = "active.vault.service.test.consul"`,
		},
		{
			name:     "test template func: grantCredentials",
			template: `[[ grantCredentials "my-db" "full" ]]`,
			wantTemplate: `
path "my-db/creds/full" {
  capabilities = ["read"]
}`,
		},
		{
			name:     "test template func: githubAssignTeamPolicy",
			template: `[[ githubAssignTeamPolicy "my-team" "my-policy" ]]`,
			wantTemplate: `
secret "/auth/github/map/teams/my-team" {
  value = "my-policy"
}`,
		},
		{
			name:     "test template func: ldapAssignGroupPolicy",
			template: `[[ ldapAssignGroupPolicy "my-group" "my-policy" ]]`,
			wantTemplate: `
secret "/auth/ldap/groups/my-group" {
  value = "my-policy"
}`,
		},
		{
			name:     "test template func: grantCredentialsPolicy",
			template: `[[ grantCredentialsPolicy "my-db" "full" ]]`,
			wantTemplate: `
policy "my-db-full" {
  path "my-db/creds/full" {
    capabilities = ["read"]
  }
}`,
		},
		{
			name: "test template func: scratch",
			template: `[[ scratch.Set "foo" "bar" ]]
test = "[[ scratch.Get "foo" ]]"
`,
			wantTemplate: `test = "bar"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := newRenderer(nil, nil)
			renderer.variables = tt.templateVariables
			require.NoError(t, err)

			got, err := renderer.renderContent(tt.template, "test", 0)
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
