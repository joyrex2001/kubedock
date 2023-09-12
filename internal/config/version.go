package config

import (
	"fmt"
)

var (
	// Version as injected during buildtime.
	Version = "<undef>"
	// Build id asinjected during buildtime.
	Build = "<undef>"
	// Date of build as injected during buildtime.
	Date = "<undef>"
	// Image is the current image as injected during buildtime.
	Image = "joyrex2001/kubedock:latest"
)

// VersionString will return a string with details of the current version.
func VersionString() string {
	return fmt.Sprintf("kubedock %s (%s)", Version, Date)
}
