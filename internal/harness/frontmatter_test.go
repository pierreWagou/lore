package harness_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pierreWagou/lore/internal/harness"
)

func TestParseFrontmatterPresent(t *testing.T) {
	content := []byte("---\nname: \"my-skill\"\ndescription: \"Does something\"\n---\n\n# My Skill\n\nContent here.\n")
	meta, body := harness.ParseFrontmatter(content)

	assert.Equal(t, "my-skill", meta.Name)
	assert.Equal(t, "Does something", meta.Description)
	assert.Equal(t, "# My Skill\n\nContent here.\n", string(body))
}

func TestParseFrontmatterAbsent(t *testing.T) {
	content := []byte("# My Skill\n\nNo frontmatter here.\n")
	meta, body := harness.ParseFrontmatter(content)

	assert.Empty(t, meta.Name)
	assert.Equal(t, content, body)
}

func TestParseFrontmatterWithGlobs(t *testing.T) {
	content := []byte("---\nname: \"ts-skill\"\nglobs:\n  - \"**/*.ts\"\nalwaysApply: true\n---\n\nContent.\n")
	meta, _ := harness.ParseFrontmatter(content)

	assert.Equal(t, "ts-skill", meta.Name)
	assert.Equal(t, []string{"**/*.ts"}, meta.Globs)
	assert.True(t, meta.AlwaysApply)
}

func TestStripFrontmatter(t *testing.T) {
	content := []byte("---\nname: \"x\"\n---\n\n# Body\n")
	body := harness.StripFrontmatter(content)
	assert.Equal(t, "# Body\n", string(body))
}

func TestStripFrontmatterNoOp(t *testing.T) {
	content := []byte("# No frontmatter\n")
	body := harness.StripFrontmatter(content)
	assert.Equal(t, content, body)
}

func TestNeedsTransform(t *testing.T) {
	assert.False(t, harness.Get("opencode").NeedsTransform())
	assert.False(t, harness.Get("claude").NeedsTransform())
}
