//go:build !integration

package bootstrap

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFlux_createHelmRepositoryManifest(t *testing.T) {
	// GIVEN
	mockCmd, f := setupFlux(t)

	mockCmd.EXPECT().RunWithOutput(
		"flux", "create", "source", "helm", "helm-repository-name", "--export",
		"-n=helm-repository-namespace", "--url=https://charts.gitlab.io").
		Return([]byte("content"), nil)

	actualFile, err := f.createHelmRepositoryManifest()

	// THEN
	require.NoError(t, err)
	assert.Equal(t, "manifest-path/helm-repository-filepath", actualFile.path)
	assert.Equal(t, actualFile.content, []byte("content"))
}

func TestFlux_createHelmRepositoryManifest_Failure(t *testing.T) {
	// GIVEN
	mockCmd, f := setupFlux(t)

	mockCmd.EXPECT().RunWithOutput(
		"flux", "create", "source", "helm", "helm-repository-name", "--export",
		"-n=helm-repository-namespace", "--url=https://charts.gitlab.io").
		Return(nil, errors.New("test"))

	actualFile, err := f.createHelmRepositoryManifest()

	// THEN
	require.Error(t, err)
	assert.Equal(t, file{}, actualFile)
}

func TestFlux_createHelmReReleaseManifest(t *testing.T) {
	// GIVEN
	mockCmd, f := setupFlux(t)

	mockCmd.EXPECT().RunWithOutput(
		"flux", "create", "helmrelease", "helm-release-name", "--export",
		"-n=helm-release-namespace", "--target-namespace=helm-release-target-namespace",
		"--create-target-namespace=true", "--source=HelmRepository/helm-repository-name.helm-repository-namespace",
		"--chart=gitlab-agent", "--release-name=helm-release-name", StartsWith("--values="), "--values=helm-release-values-1", "--values=helm-release-values-2", "--values-from=helm-release-values-from-1").
		Return([]byte("content"), nil)

	actualFile, err := f.createHelmReleaseManifest("wss://kas.gitlab.example.com")

	// THEN
	require.NoError(t, err)
	assert.Equal(t, "manifest-path/helm-release-filepath", actualFile.path)
	assert.Equal(t, actualFile.content, []byte("content"))
}

func TestFlux_createHelmReReleaseManifest_Failure(t *testing.T) {
	// GIVEN
	mockCmd, f := setupFlux(t)

	mockCmd.EXPECT().RunWithOutput(
		"flux", "create", "helmrelease", "helm-release-name", "--export",
		"-n=helm-release-namespace", "--target-namespace=helm-release-target-namespace",
		"--create-target-namespace=true", "--source=HelmRepository/helm-repository-name.helm-repository-namespace",
		"--chart=gitlab-agent", "--release-name=helm-release-name", StartsWith("--values="), "--values=helm-release-values-1", "--values=helm-release-values-2", "--values-from=helm-release-values-from-1").
		Return([]byte(""), errors.New("test"))

	actualFile, err := f.createHelmReleaseManifest("wss://kas.gitlab.example.com")

	// THEN
	require.Error(t, err)
	assert.Equal(t, file{}, actualFile)
}

func TestFlux_reconcile(t *testing.T) {
	// GIVEN
	mockCmd, f := setupFlux(t)

	gomock.InOrder(
		mockCmd.EXPECT().Run("flux", "reconcile", "source", "flux-source-type", "flux-source-name", "-n=flux-source-namespace"),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace"),
		mockCmd.EXPECT().Run("flux", "reconcile", "helmrelease", "helm-release-name", "-n=helm-release-namespace", "--with-source"),
	)

	// WHEN
	_ = f.reconcile()
}

func TestFlux_reconcile_retries(t *testing.T) {
	// GIVEN
	mockCmd, f := setupFlux(t)

	gomock.InOrder(
		mockCmd.EXPECT().Run("flux", "reconcile", "source", "flux-source-type", "flux-source-name", "-n=flux-source-namespace"),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace").Return([]byte(`HelmRelease object 'helm-release-name' not found in "helm-release-namespace" namespace`), nil),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace"),
		mockCmd.EXPECT().Run("flux", "reconcile", "helmrelease", "helm-release-name", "-n=helm-release-namespace", "--with-source"),
	)

	// WHEN
	_ = f.reconcile()
}

func TestFlux_reconcile_abort_retry_max(t *testing.T) {
	// GIVEN
	mockCmd, f := setupFlux(t)

	gomock.InOrder(
		mockCmd.EXPECT().Run("flux", "reconcile", "source", "flux-source-type", "flux-source-name", "-n=flux-source-namespace"),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace").Return([]byte(`HelmRelease object 'helm-release-name' not found in "helm-release-namespace" namespace`), nil),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace").Return([]byte(`HelmRelease object 'helm-release-name' not found in "helm-release-namespace" namespace`), nil),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace").Return([]byte(`HelmRelease object 'helm-release-name' not found in "helm-release-namespace" namespace`), nil),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace").Return([]byte(`HelmRelease object 'helm-release-name' not found in "helm-release-namespace" namespace`), nil),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace").Return([]byte(`HelmRelease object 'helm-release-name' not found in "helm-release-namespace" namespace`), nil),
		mockCmd.EXPECT().RunWithOutput("flux", "get", "helmreleases", "helm-release-name", "-n=helm-release-namespace").Return([]byte(`HelmRelease object 'helm-release-name' not found in "helm-release-namespace" namespace`), nil),
	)

	// WHEN
	err := f.reconcile()

	assert.Error(t, err)
}

func setupFlux(t *testing.T) (*MockCmd, FluxWrapper) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockCmd := NewMockCmd(ctrl)
	f := NewLocalFluxWrapper(
		mockCmd,
		"flux", "manifest-path",
		"https://charts.gitlab.io",
		"helm-repository-name", "helm-repository-namespace", "helm-repository-filepath",
		"helm-release-name", "helm-release-namespace", "helm-release-filepath", "helm-release-target-namespace",
		[]string{"helm-release-values-1", "helm-release-values-2"}, []string{"helm-release-values-from-1"},
		"flux-source-type", "flux-source-namespace", "flux-source-name",
	)
	fHack, ok := f.(*localFluxWrapper)
	require.True(t, ok, "newFlux should return a *localFluxWrapper in tests")
	fHack.reconcileRetryDelay = 0

	return mockCmd, f
}

func StartsWith(prefix string) gomock.Matcher {
	return &startsWithMatcher{prefix: prefix}
}

type startsWithMatcher struct {
	prefix  string
	actualS string
}

func (m startsWithMatcher) Matches(arg any) bool {
	s, ok := arg.(string)
	if !ok {
		return false
	}
	m.actualS = s
	return strings.HasPrefix(m.actualS, m.prefix)
}

func (m startsWithMatcher) String() string {
	return fmt.Sprintf("does not start with: %q, got %q", m.prefix, m.actualS)
}
