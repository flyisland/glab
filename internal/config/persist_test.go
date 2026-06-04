//go:build !integration

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An in-memory config (no directory behind it) must not persist to disk, even
// though Write()/WriteAll() are called. This is the misuse the test seams used
// to work around with StubWriteConfig / noWriteConfig.
func Test_InMemoryConfig_DoesNotPersist(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GLAB_CONFIG_DIR", dir)

	cfg := NewBlankConfig()
	require.NoError(t, cfg.Set("", "git_protocol", "https"))
	require.NoError(t, cfg.WriteAll())

	_, err := os.Stat(filepath.Join(dir, "config.yml"))
	assert.True(t, os.IsNotExist(err), "in-memory config must not persist to disk")
}

// A dir-backed config persists seeded and mutated values to that directory,
// across both the main config file and the aliases file.
func Test_NewFromStringInDir_Persists(t *testing.T) {
	dir := t.TempDir()

	cfg := NewFromStringInDir("git_protocol: ssh\n", dir)
	require.NoError(t, cfg.Set("", "editor", "vim"))

	aliases, err := cfg.Aliases()
	require.NoError(t, err)
	require.NoError(t, aliases.Set("co", "mr checkout"))

	require.NoError(t, cfg.WriteAll())

	assert.Contains(t, persistedFile(t, dir, "config.yml"), "editor: vim")
	assert.Contains(t, persistedFile(t, dir, "aliases.yml"), "co: mr checkout")
}

// Regression guard for the first-run path: Init() parses a not-yet-existing
// config file and then WriteAll()s the defaults. ParseConfig must remember the
// source directory even when the file is missing, or first-run seeding silently
// stops writing.
func Test_ParseConfig_FirstRun_SeedsDefaultsToSourceDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GLAB_CONFIG_DIR", dir)

	cfg, err := ParseConfig(filepath.Join(dir, "config.yml"))
	require.True(t, err == nil || os.IsNotExist(err))

	require.NoError(t, cfg.WriteAll())

	_, statErr := os.Stat(filepath.Join(dir, "config.yml"))
	require.NoError(t, statErr, "first-run config must be seeded to disk on WriteAll")
}
