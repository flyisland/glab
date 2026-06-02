//go:build !integration

package securefile

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/cli/internal/testing/cmdtest"
)

func Test_Securefile(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCmdSecurefile(cmdtest.NewTestFactory(nil))
	cmd.SetOut(&buf)

	require.NoError(t, cmd.Execute())

	assert.Contains(t, buf.String(), "Use \"securefile [command] --help\" for more information about a command.\n")
}
