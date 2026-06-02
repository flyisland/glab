//go:build !integration

package iteration

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/cli/internal/testing/cmdtest"
)

func TestNewCmdIteration(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCmdIteration(cmdtest.NewTestFactory(nil))
	cmd.SetOut(&buf)

	require.NoError(t, cmd.Execute())

	assert.Contains(t, buf.String(), "Use \"iteration [command] --help\" for more information about a command.\n")
}
