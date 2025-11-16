package version

import (
	"os/exec"
	"strings"
)

var (
	// These will be set at build time or runtime
	Version   string
	CommitID  string
	BuildTime string
)

func init() {
	// Get version from git tag
	if Version == "" {
		cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
		if output, err := cmd.Output(); err == nil {
			Version = strings.TrimSpace(string(output))
		} else {
			Version = "dev"
		}
	}

	// Get short commit ID
	if CommitID == "" {
		cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
		if output, err := cmd.Output(); err == nil {
			CommitID = strings.TrimSpace(string(output))
		} else {
			CommitID = "unknown"
		}
	}
}

// GetVersion returns the version string
func GetVersion() string {
	if Version == "dev" || Version == "" {
		return "dev"
	}
	return Version
}

// GetCommitID returns the short commit ID
func GetCommitID() string {
	return CommitID
}

// GetFullVersion returns version with commit ID
func GetFullVersion() string {
	if Version == "dev" || Version == "" {
		return "dev-" + CommitID
	}
	return Version + "-" + CommitID
}
