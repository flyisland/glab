//go:build !integration

package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/cli/internal/dependencyfirewall/verdict"
)

func TestClassify403IsBlockedWithReasonFromBody(t *testing.T) {
	body := []byte(`{"error":"blocked by policy: known malware"}`)
	e, ok := classify("foo", "1.2.3", 403, nil, body)
	assert.True(t, ok)
	assert.Equal(t, verdict.Blocked, e.Verdict)
	assert.Equal(t, 403, e.Status)
	assert.Contains(t, e.Reason, "known malware")
}

func TestClassify403PrefersMessageOverError(t *testing.T) {
	body := []byte(`{"message":"Package 'lodash' violates 'block-mit-npm' policy","error":"Dependency Firewall policy violation"}`)
	e, ok := classify("lodash", "4.17.21", 403, nil, body)
	assert.True(t, ok)
	assert.Equal(t, verdict.Blocked, e.Verdict)
	assert.Equal(t, "Package 'lodash' violates 'block-mit-npm' policy", e.Reason)
}

func TestClassify403WithUnparseableBodyUsesGenericReason(t *testing.T) {
	e, ok := classify("foo", "1.2.3", 403, nil, []byte("not json"))
	assert.True(t, ok)
	assert.Equal(t, verdict.Blocked, e.Verdict)
	assert.NotEmpty(t, e.Reason)
}

func TestClassifyNon403Non2xxIsNotRecorded(t *testing.T) {
	for _, status := range []int{404, 500, 502} {
		_, ok := classify("foo", "1.2.3", status, nil, nil)
		assert.False(t, ok, "status %d should not be a verdict", status)
	}
}

func TestClassify2xxWithWarningHeaderIsWarning(t *testing.T) {
	h := map[string][]string{firewallWarningHeader: {"license review required"}}
	e, ok := classify("bar", "0.9.0", 200, h, nil)
	assert.True(t, ok)
	assert.Equal(t, verdict.Warning, e.Verdict)
	assert.Contains(t, e.Reason, "license review")
}

func TestClassify2xxNoSignalIsNotRecorded(t *testing.T) {
	_, ok := classify("baz", "1.0.0", 200, nil, []byte(`{"name":"baz"}`))
	assert.False(t, ok)
}
