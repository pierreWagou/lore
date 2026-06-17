package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendEntriesToFileCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	require.NoError(t, appendEntriesToFile(path, "# lore", []string{".opencode/skills/"}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), ".opencode/skills/")
	assert.Contains(t, string(content), "# lore")
}

func TestAppendEntriesToFileIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	require.NoError(t, appendEntriesToFile(path, "# lore", []string{".opencode/skills/"}))
	require.NoError(t, appendEntriesToFile(path, "# lore", []string{".opencode/skills/"}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	// The entry should appear exactly once.
	count := 0
	for _, line := range splitLines(string(content)) {
		if line == ".opencode/skills/" {
			count++
		}
	}
	assert.Equal(t, 1, count, "entry should appear exactly once")
}

func TestAppendEntriesToFileDoesNotFalsePositiveOnSubstring(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	// Pre-populate file with a comment that contains the entry text as a substring.
	require.NoError(t, os.WriteFile(path, []byte("# .opencode/skills/ goes here\n"), 0644))

	require.NoError(t, appendEntriesToFile(path, "# lore", []string{".opencode/skills/"}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	// The entry should now be present as its own line.
	found := false
	for _, line := range splitLines(string(content)) {
		if line == ".opencode/skills/" {
			found = true
			break
		}
	}
	assert.True(t, found, "entry should be added even when it appears as a substring in a comment")
}

func TestAppendEntriesToFileAppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	require.NoError(t, os.WriteFile(path, []byte("node_modules/\n"), 0644))
	require.NoError(t, appendEntriesToFile(path, "# lore", []string{".opencode/skills/"}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "node_modules/")
	assert.Contains(t, string(content), ".opencode/skills/")
}

func TestContainsLine(t *testing.T) {
	assert.True(t, containsLine("a\nb\nc\n", "b"))
	assert.False(t, containsLine("a\nbc\nc\n", "b"))
	assert.False(t, containsLine("# b is here\n", "b"))
	assert.True(t, containsLine("b\n", "b"))
	assert.True(t, containsLine("b", "b"))
}

// splitLines splits a string by newline for test assertions.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
