package params

import (
	"fmt"
)

const (
	VersionMajor = 1        // Major version component of the current release
	VersionMinor = 3        // Minor version component of the current release
	VersionPatch = 0        // Patch version component of the current release
	VersionMeta  = "stable" // Version metadata to append to the version string
)

type VersionInfo struct {
	Major uint64
	Minor uint64
	Patch uint64
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func cmp(x uint64, y uint64) int {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

func (v *VersionInfo) Cmp(version *VersionInfo) int {
	if v.Major == version.Major {
		if v.Minor == version.Minor {
			return cmp(v.Patch, version.Patch)
		}
		return cmp(v.Minor, version.Minor)
	}
	return cmp(v.Major, version.Major)
}

var CurrentVersionInfo = func() *VersionInfo {
	return &VersionInfo{VersionMajor, VersionMinor, VersionPatch}
}()

// Version holds the textual version string.
var Version = func() string {
	return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
}()

// VersionWithMeta holds the textual version string including the metadata.
var VersionWithMeta = func() string {
	v := Version
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}()

// ArchiveVersion holds the textual version string used for Atlas archives.
// e.g. "1.8.11-dea1ce05" for stable releases, or
//
//	"1.8.13-unstable-21c059b6" for unstable releases
func ArchiveVersion(gitCommit string) string {
	vsn := Version
	if VersionMeta != "stable" {
		vsn += "-" + VersionMeta
	}
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	return vsn
}

func VersionWithCommit(gitCommit, gitDate string) string {
	vsn := VersionWithMeta
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	if (VersionMeta != "stable") && (gitDate != "") {
		vsn += "-" + gitDate
	}
	return vsn
}
