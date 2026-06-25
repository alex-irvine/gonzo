package version

import (
	"strings"
)

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	UpdateAvailable bool
	LatestVersion   string
	CurrentVersion  string
	ReleaseURL      string
	Severity        string
}

// Checker tracks the running build's version information.
//
// The automatic update check was removed: it pointed at the upstream
// gonzo-version.controltheory.com service, which only knows about the
// original ControlTheory/gonzo releases — not this fork. It reported
// upstream versions and steered users at the wrong repo. The footer now
// just shows the current build version (derived from git tags via ldflags).
type Checker struct {
	currentVersion string
	commit         string
}

// NewChecker creates a new version checker
func NewChecker(currentVersion, commit string) *Checker {
	return &Checker{
		currentVersion: currentVersion,
		commit:         commit,
	}
}

// CheckInBackground is a no-op. The remote update check was removed; see the
// Checker doc comment.
func (c *Checker) CheckInBackground() {}

// GetUpdateInfo returns nil; no remote update check is performed.
func (c *Checker) GetUpdateInfo() *UpdateInfo { return nil }

// GetUpdateInfoNonBlocking returns nil; no remote update check is performed.
func (c *Checker) GetUpdateInfoNonBlocking() *UpdateInfo { return nil }

// GetCurrentVersion returns the running build's version, without a leading "v".
func (c *Checker) GetCurrentVersion() string {
	return strings.TrimPrefix(c.currentVersion, "v")
}
