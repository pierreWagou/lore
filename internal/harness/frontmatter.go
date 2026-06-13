package harness

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMeta is the optional YAML frontmatter block in a SKILL.md file.
// It follows the agentskills.io open standard.
type SkillMeta struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Globs       []string `yaml:"globs"`
	AlwaysApply bool     `yaml:"alwaysApply"`
}

// ParseFrontmatter extracts YAML frontmatter from markdown content.
// Returns the parsed metadata and the body without the frontmatter block.
// If no frontmatter is present, meta is zero-value and body equals the original content.
func ParseFrontmatter(content []byte) (SkillMeta, []byte) {
	s := string(content)
	if !strings.HasPrefix(s, "---\n") {
		return SkillMeta{}, content
	}

	rest := s[4:] // skip opening "---\n"
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return SkillMeta{}, content
	}

	yamlBlock := rest[:idx]
	after := rest[idx+4:] // skip "\n---"
	body := strings.TrimLeft(after, "\n")

	var meta SkillMeta
	_ = yaml.Unmarshal([]byte(yamlBlock), &meta)
	return meta, []byte(body)
}

// StripFrontmatter returns the content with the YAML frontmatter block removed.
func StripFrontmatter(content []byte) []byte {
	_, body := ParseFrontmatter(content)
	return body
}
