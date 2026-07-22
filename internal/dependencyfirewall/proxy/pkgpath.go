package proxy

import (
	"net/url"
	"strings"
)

func packageFromPath(rawPath string) (string, string, bool) {
	decoded, err := url.PathUnescape(rawPath)
	if err != nil {
		decoded = rawPath
	}
	trimmed := strings.Trim(stripRegistryPrefix(decoded), "/")
	if trimmed == "" {
		return "", "", false
	}

	parts := strings.Split(trimmed, "/")
	if parts[0] == "-" {
		return "", "", false
	}

	var name string
	var rest []string
	if strings.HasPrefix(parts[0], "@") {
		if len(parts) < 2 {
			return "", "", false
		}
		name = parts[0] + "/" + parts[1]
		rest = parts[2:]
	} else {
		name = parts[0]
		rest = parts[1:]
	}

	if len(rest) == 0 {
		return name, "", true
	}

	if rest[0] == "-" {
		if v := versionFromTarball(rest, name); v != "" {
			return name, v, true
		}
		return name, "", true
	}

	return name, rest[0], true
}

// stripRegistryPrefix removes a GitLab npm registry path prefix
// (for example, "api/v4/projects/<id>/packages/npm/" or
// "api/v4/packages/npm/") so the remainder begins with the npm package
// coordinates. Paths without the prefix (such as a direct npmjs.org request)
// are returned unchanged.
func stripRegistryPrefix(path string) string {
	const marker = "packages/npm/"
	if i := strings.LastIndex(path, marker); i >= 0 {
		return path[i+len(marker):]
	}
	return path
}

func versionFromTarball(rest []string, name string) string {
	last := rest[len(rest)-1]
	if !strings.HasSuffix(last, ".tgz") {
		return ""
	}
	base := strings.TrimSuffix(last, ".tgz")
	short := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		short = name[i+1:]
	}
	return strings.TrimPrefix(base, short+"-")
}
