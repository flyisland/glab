package proxy

import (
	"encoding/json"
	"net/http"
	"strings"

	"gitlab.com/gitlab-org/cli/internal/dependencyfirewall/verdict"
)

const firewallWarningHeader = "X-Gitlab-Dependency-Firewall-Warning"

var blockedReasonFields = []string{"message", "error", "reason"}

func classify(pkg, version string, status int, header http.Header, body []byte) (verdict.Entry, bool) {
	switch {
	case status == http.StatusForbidden:
		return verdict.Entry{
			Package: pkg,
			Version: version,
			Verdict: verdict.Blocked,
			Status:  status,
			Reason:  blockedReason(body),
		}, true
	case status >= 200 && status < 300:
		if reason, ok := warningReason(header, body); ok {
			return verdict.Entry{
				Package: pkg,
				Version: version,
				Verdict: verdict.Warning,
				Status:  status,
				Reason:  reason,
			}, true
		}
		return verdict.Entry{}, false
	default:
		return verdict.Entry{}, false
	}
}

func blockedReason(body []byte) string {
	if r := fieldFromJSON(body, blockedReasonFields); r != "" {
		return r
	}
	return "Package blocked by GitLab Dependency Firewall policy."
}

func warningReason(header http.Header, body []byte) (string, bool) {
	if header != nil {
		if v := header.Get(firewallWarningHeader); v != "" {
			return v, true
		}
	}
	if r := fieldFromJSON(body, []string{"warning", "dependencyFirewallWarning"}); r != "" {
		return r, true
	}
	return "", false
}

func fieldFromJSON(body []byte, fields []string) string {
	if len(body) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return ""
	}
	for _, f := range fields {
		if v, ok := m[f]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}
	return ""
}
