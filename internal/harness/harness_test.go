package harness_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pierreWagou/lore/internal/harness"
)

func TestGetKnownHarness(t *testing.T) {
	a := harness.Get("opencode")
	require.NotNil(t, a)
	assert.Equal(t, "opencode", a.Name())

	b := harness.Get("claude")
	require.NotNil(t, b)
	assert.Equal(t, "claude", b.Name())
}

func TestGetUnknownHarness(t *testing.T) {
	assert.Nil(t, harness.Get("notaharness"))
}

func TestAllHarnesses(t *testing.T) {
	all := harness.All()
	assert.Len(t, all, 2)
}

func TestNames(t *testing.T) {
	names := harness.Names()
	assert.Contains(t, names, "opencode")
	assert.Contains(t, names, "claude")
}

func TestOpenCodeTransform(t *testing.T) {
	a := harness.Get("opencode")
	skill := harness.Skill{
		Name:  "my-skill",
		Files: map[string][]byte{"SKILL.md": []byte("# my skill")},
	}
	files, err := a.Transform(skill)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "SKILL.md", files[0].Path)
	assert.Equal(t, []byte("# my skill"), files[0].Content)
}

func TestOpenCodeTransformMissingSkillMD(t *testing.T) {
	a := harness.Get("opencode")
	skill := harness.Skill{
		Name:  "my-skill",
		Files: map[string][]byte{"README.md": []byte("readme")},
	}
	_, err := a.Transform(skill)
	assert.Error(t, err)
}

func TestClaudeTransform(t *testing.T) {
	a := harness.Get("claude")
	skill := harness.Skill{
		Name:  "my-skill",
		Files: map[string][]byte{"SKILL.md": []byte("# my skill")},
	}
	files, err := a.Transform(skill)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "SKILL.md", files[0].Path)
}

func TestOpenCodeSkillsDir(t *testing.T) {
	a := harness.Get("opencode")
	dir := a.GlobalSkillsDir()
	assert.Contains(t, dir, "opencode")
	assert.Contains(t, dir, "skills")

	projDir := a.ProjectSkillsDir("/my/project")
	assert.Equal(t, "/my/project/.opencode/skills", projDir)
}

func TestClaudeSkillsDir(t *testing.T) {
	a := harness.Get("claude")
	dir := a.GlobalSkillsDir()
	assert.Contains(t, dir, ".claude")
	assert.Contains(t, dir, "skills")

	projDir := a.ProjectSkillsDir("/my/project")
	assert.Equal(t, "/my/project/.claude/skills", projDir)
}
