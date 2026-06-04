//go:build !integration

package glrepo

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/internal/glinstance"
)

// FromURL must use the injected config (not a global) to strip a configured
// subfolder from the URL path.
func Test_FromURL_UsesInjectedConfigForSubfolder(t *testing.T) {
	t.Parallel()

	cfg := config.NewFromString(`
hosts:
  gitlab.example.com:
    subfolder: gitlab
`)

	u, err := url.Parse("https://gitlab.example.com/gitlab/group/repo")
	require.NoError(t, err)

	repo, err := FromURL(u, glinstance.DefaultHostname, cfg)
	require.NoError(t, err)
	assert.Equal(t, "group", repo.RepoOwner())
	assert.Equal(t, "repo", repo.RepoName())
}

// FromFullName must use the injected config's host list (not a global) to treat
// a dotted first segment as a known host rather than a group name.
func Test_FromFullName_UsesInjectedConfigForHosts(t *testing.T) {
	t.Parallel()

	cfg := config.NewFromString(`
hosts:
  my.host.com:
    token: TOKEN
`)

	repo, err := FromFullName("my.host.com/owner/repo", glinstance.DefaultHostname, cfg)
	require.NoError(t, err)
	assert.Equal(t, "my.host.com", repo.RepoHost())
	assert.Equal(t, "owner", repo.RepoOwner())
	assert.Equal(t, "repo", repo.RepoName())
}

// A nil config must be tolerated (config-unavailable paths behave as before).
func Test_FromURL_NilConfig(t *testing.T) {
	t.Parallel()

	u, err := url.Parse("https://gitlab.com/group/repo")
	require.NoError(t, err)

	repo, err := FromURL(u, glinstance.DefaultHostname, nil)
	require.NoError(t, err)
	assert.Equal(t, "group", repo.RepoOwner())
	assert.Equal(t, "repo", repo.RepoName())
}
