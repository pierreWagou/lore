package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pierreWagou/lore/internal/resolver"
)

// --- inferName ---

func TestInferNameFromSubPath(t *testing.T) {
	assert.Equal(t, "standup", inferName("skills/standup", "owner/repo"))
}

func TestInferNameFromFallback(t *testing.T) {
	assert.Equal(t, "my-skill", inferName("", "owner/repo/my-skill"))
}

func TestInferNameFromFallbackShort(t *testing.T) {
	assert.Equal(t, "repo", inferName("", "owner/repo"))
}

// --- buildSource ---

func TestBuildSourceShorthand(t *testing.T) {
	h := resolver.Handle{
		Kind:     resolver.KindShorthand,
		Owner:    "alan-eu",
		RepoName: "alan-skills",
	}
	assert.Equal(t, "alan-eu/alan-skills/skills/standup", buildSource(h, "skills/standup"))
}

func TestBuildSourceHTTPS(t *testing.T) {
	h := resolver.Handle{
		Kind:    resolver.KindHTTPS,
		RepoURL: "https://github.com/alan-eu/alan-skills",
	}
	assert.Equal(t, "https://github.com/alan-eu/alan-skills/skills/standup", buildSource(h, "skills/standup"))
}

func TestBuildSourceSSH(t *testing.T) {
	h := resolver.Handle{
		Kind:    resolver.KindSSH,
		RepoURL: "git@github.com:alan-eu/alan-skills.git",
	}
	assert.Equal(t, "git@github.com:alan-eu/alan-skills/skills/standup", buildSource(h, "skills/standup"))
}

// --- splitHarnesses ---

func TestSplitHarnessesEmpty(t *testing.T) {
	assert.Nil(t, splitHarnesses(""))
}

func TestSplitHarnessesSingle(t *testing.T) {
	assert.Equal(t, []string{"opencode"}, splitHarnesses("opencode"))
}

func TestSplitHarnessesMultiple(t *testing.T) {
	assert.Equal(t, []string{"opencode", "claude"}, splitHarnesses("opencode,claude"))
}

func TestSplitHarnessesTrimsSpaces(t *testing.T) {
	assert.Equal(t, []string{"opencode", "claude"}, splitHarnesses(" opencode , claude "))
}

func TestSplitHarnessesSkipsEmpty(t *testing.T) {
	assert.Equal(t, []string{"opencode"}, splitHarnesses("opencode,,"))
}

// --- maskToken ---

func TestMaskTokenShort(t *testing.T) {
	assert.Equal(t, "****", maskToken("short"))
	assert.Equal(t, "****", maskToken("12345678"))
}

func TestMaskTokenLong(t *testing.T) {
	assert.Equal(t, "ghp_...oken", maskToken("ghp_averylongtoken"))
}

// --- removeFromSlice ---

func TestRemoveFromSliceFound(t *testing.T) {
	s := []string{"a", "b", "c"}
	removed := removeFromSlice(&s, "b")
	assert.True(t, removed)
	assert.Equal(t, []string{"a", "c"}, s)
}

func TestRemoveFromSliceFirst(t *testing.T) {
	s := []string{"a", "b", "c"}
	removed := removeFromSlice(&s, "a")
	assert.True(t, removed)
	assert.Equal(t, []string{"b", "c"}, s)
}

func TestRemoveFromSliceLast(t *testing.T) {
	s := []string{"a", "b", "c"}
	removed := removeFromSlice(&s, "c")
	assert.True(t, removed)
	assert.Equal(t, []string{"a", "b"}, s)
}

func TestRemoveFromSliceNotFound(t *testing.T) {
	s := []string{"a", "b"}
	removed := removeFromSlice(&s, "z")
	assert.False(t, removed)
	assert.Equal(t, []string{"a", "b"}, s)
}

func TestRemoveFromSliceEmpty(t *testing.T) {
	var s []string
	removed := removeFromSlice(&s, "x")
	assert.False(t, removed)
}
