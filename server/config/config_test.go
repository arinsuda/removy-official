package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAllowedOrigins(t *testing.T) {
	got, err := parseAllowedOrigins(" https://a.example/, http://localhost:3000,https://a.example,, ")
	require.NoError(t, err)
	assert.Equal(t, []string{"https://a.example", "http://localhost:3000"}, got)

	for _, raw := range []string{"", " , ", "*", "https://*.example.com/path", "a.example", "ftp://a.example", "https://a.example/app"} {
		_, err := parseAllowedOrigins(raw)
		assert.Error(t, err, raw)
	}
}

func TestCORSConfigIsAllowed(t *testing.T) {
	c := CORSConfig{AllowedOrigins: []string{"https://a.example"}}
	assert.True(t, c.IsAllowed("https://a.example"))
	for _, origin := range []string{"", "null", "https://evil.example", "https://a.example.evil.example", "http://a.example"} {
		assert.False(t, c.IsAllowed(origin), origin)
	}
}

func TestLoadDotEnvPrecedence(t *testing.T) {
	root := t.TempDir()
	server := filepath.Join(root, "server")
	require.NoError(t, os.Mkdir(server, 0o755))
	write := func(path, content string) {
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	write(filepath.Join(root, ".env"), "DOTENV_T_REAL=root\nDOTENV_T_SERVER=root\nDOTENV_T_ROOT=root\n")
	write(filepath.Join(server, ".env"), "DOTENV_T_REAL=server\nDOTENV_T_SERVER=server\n")
	t.Chdir(server)
	t.Setenv("DOTENV_T_REAL", "real")
	for _, key := range []string{"DOTENV_T_SERVER", "DOTENV_T_ROOT"} {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}

	require.NoError(t, LoadDotEnv())
	assert.Equal(t, "real", os.Getenv("DOTENV_T_REAL"))
	assert.Equal(t, "server", os.Getenv("DOTENV_T_SERVER"))
	assert.Equal(t, "root", os.Getenv("DOTENV_T_ROOT"))
}

func TestLoadDotEnvMissingFilesIsNotAnError(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, LoadDotEnv())
}
