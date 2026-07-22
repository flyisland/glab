//go:build !integration

package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageFromPath(t *testing.T) {
	cases := []struct {
		path        string
		wantPkg     string
		wantVersion string
		wantOK      bool
	}{
		{"/lodash", "lodash", "", true},
		{"/lodash/4.17.21", "lodash", "4.17.21", true},
		{"/lodash/-/lodash-4.17.21.tgz", "lodash", "4.17.21", true},
		{"/@types%2fnode", "@types/node", "", true},
		{"/@types/node", "@types/node", "", true},
		{"/@types/node/-/node-20.1.0.tgz", "@types/node", "20.1.0", true},
		{"/", "", "", false},
		{"/-/ping", "", "", false},
		// GitLab project-level registry paths carry an /api/v4/projects/<id>/packages/npm/ prefix.
		{"/api/v4/projects/26/packages/npm/lodash", "lodash", "", true},
		{"/api/v4/projects/26/packages/npm/lodash/-/lodash-4.17.21.tgz", "lodash", "4.17.21", true},
		{"/api/v4/projects/26/packages/npm/@types/node/-/node-20.1.0.tgz", "@types/node", "20.1.0", true},
		{"/api/v4/projects/group%2Fproj/packages/npm/lodash/-/lodash-4.17.21.tgz", "lodash", "4.17.21", true},
		// GitLab group/instance-level registry paths.
		{"/api/v4/packages/npm/lodash/-/lodash-4.17.21.tgz", "lodash", "4.17.21", true},
		// Prefix present but no package after it.
		{"/api/v4/projects/26/packages/npm/", "", "", false},
	}
	for _, c := range cases {
		pkg, version, ok := packageFromPath(c.path)
		assert.Equal(t, c.wantOK, ok, c.path)
		assert.Equal(t, c.wantPkg, pkg, c.path)
		assert.Equal(t, c.wantVersion, version, c.path)
	}
}
