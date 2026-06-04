//go:build !integration

package logout

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/internal/testing/cmdtest"
)

func Test_NewCmdLogout(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		stdErr   string
		wantErr  bool
	}{
		{
			name:     "no arguments",
			hostname: "",
			stdErr:   "hostname is required to logout. Use --hostname flag to specify hostname",
			wantErr:  true,
		},
		{
			name:     "hostname set",
			hostname: "gitlab.example.com",
			stdErr:   "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			token := "xxxxxxxx"

			cfg := config.NewFromStringInDir(heredoc.Docf(
				`
					hosts:
					  gitlab.something.com:
					    token: %[1]s
					  gitlab.example.com:
					    token: %[1]s
					    job_token: %[1]s
					    is_oauth2: %[1]s
					    oauth2_refresh_token: %[1]s
					    oauth2_expiry_date: %[1]s
				`,
				token,
			), dir)

			// removing the environment variable so CI does not interfere
			t.Setenv("GITLAB_TOKEN", "")

			exec := cmdtest.SetupCmdForTest(t, NewCmdLogout, true, cmdtest.WithConfig(cfg))
			output, err := exec(fmt.Sprintf("--hostname %s", tt.hostname))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				logoutMessage := fmt.Sprintf("Successfully logged out of %s\n", tt.hostname)
				assert.Equal(t, logoutMessage, output.String())

				data, err := os.ReadFile(filepath.Join(dir, "config.yml"))
				require.NoError(t, err)
				cfg := config.NewFromString(string(data))
				gitlabToken, _ := cfg.Get("gitlab.something.com", "token")
				assert.Equal(t, token, gitlabToken)

				exampleToken, _ := cfg.Get(tt.hostname, "token")
				assert.Empty(t, exampleToken)

				exampleJobToken, _ := cfg.Get(tt.hostname, "job_token")
				assert.Empty(t, exampleJobToken)

				exampleIsOauth2, _ := cfg.Get(tt.hostname, "is_oauth2")
				assert.Empty(t, exampleIsOauth2)

				exampleOauth2RefreshToken, _ := cfg.Get(tt.hostname, "oauth2_refresh_token")
				assert.Empty(t, exampleOauth2RefreshToken)

				exampleOauth2ExpiryDate, _ := cfg.Get(tt.hostname, "oauth2_expiry_date")
				assert.Empty(t, exampleOauth2ExpiryDate)
			}
		})
	}
}
