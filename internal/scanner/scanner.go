package scanner

import "path/filepath"

const SkillFileName = "SKILL.md"

// Scan walks a file map (path → content) and returns the paths of directories
// that contain a SKILL.md file. Paths are relative to the map root.
// An empty string is returned for SKILL.md files at the root.
func Scan(files map[string][]byte) []string {
	seen := make(map[string]bool)
	var dirs []string
	for path := range files {
		if filepath.Base(path) == SkillFileName {
			dir := filepath.Dir(path)
			if dir == "." {
				dir = ""
			}
			if !seen[dir] {
				seen[dir] = true
				dirs = append(dirs, dir)
			}
		}
	}
	return dirs
}
