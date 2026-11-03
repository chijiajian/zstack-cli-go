// Package skills contains the independently installable operations Skill.
package skills

import "embed"

// Files is the version of the Skill distributed with this CLI binary.
//
//go:embed zstack-operations
var Files embed.FS
