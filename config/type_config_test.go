package config

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
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

			got, err := c.ParseContent(tt.content)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err2 := c.ProcessContent(got)
			if tt.wantErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
			}

			spew.Dump(c)
			require.Equal(t, tt.seenEnvironments, c.Environments.List())
			require.Equal(t, tt.seenApplications, c.Applications.List())
			require.Equal(t, tt.seenSecrets, c.VaultSecrets.List())
		})
	}
}
