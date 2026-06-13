package scanner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pierreWagou/lore/internal/scanner"
)

func TestScanEmpty(t *testing.T) {
	dirs := scanner.Scan(map[string][]byte{})
	assert.Empty(t, dirs)
}

func TestScanRoot(t *testing.T) {
	files := map[string][]byte{
		"SKILL.md": []byte("# skill"),
		"README.md": []byte("readme"),
	}
	dirs := scanner.Scan(files)
	assert.Equal(t, []string{""}, dirs)
}

func TestScanNested(t *testing.T) {
	files := map[string][]byte{
		"skills/pdf/SKILL.md":    []byte("# pdf"),
		"skills/web/SKILL.md":    []byte("# web"),
		"skills/web/examples.md": []byte("examples"),
		"README.md":              []byte("readme"),
	}
	dirs := scanner.Scan(files)
	assert.Len(t, dirs, 2)
	assert.Contains(t, dirs, "skills/pdf")
	assert.Contains(t, dirs, "skills/web")
}

func TestScanNoDuplicates(t *testing.T) {
	files := map[string][]byte{
		"a/SKILL.md":       []byte("# a"),
		"a/extra/SKILL.md": []byte("# a-extra"),
	}
	dirs := scanner.Scan(files)
	assert.Len(t, dirs, 2)
}

func TestHasSkill(t *testing.T) {
	assert.True(t, scanner.HasSkill(map[string][]byte{"SKILL.md": []byte("x")}))
	assert.False(t, scanner.HasSkill(map[string][]byte{"README.md": []byte("x")}))
}
