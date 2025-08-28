package utils

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

// VersionComparer provides semantic version comparison utilities
type VersionComparer struct{}

// NewVersionComparer creates a new version comparer instance
func NewVersionComparer() *VersionComparer {
	return &VersionComparer{}
}

// IsUpdateNeeded checks if an update is needed based on semantic version comparison
// Returns true if latestVersion is newer than currentVersion
func (vc *VersionComparer) IsUpdateNeeded(currentVersion, latestVersion string) bool {
	// Handle empty versions
	if currentVersion == "" || latestVersion == "" {
		return currentVersion != latestVersion
	}

	// Normalize versions (add 'v' prefix if missing for semver compatibility)
	currentVer := vc.normalizeVersion(currentVersion)
	latestVer := vc.normalizeVersion(latestVersion)

	// Try to parse as semantic versions
	current, err1 := semver.NewVersion(currentVer)
	latest, err2 := semver.NewVersion(latestVer)

	// If both versions are valid semver, use semantic comparison
	if err1 == nil && err2 == nil {
		return latest.GreaterThan(current)
	}

	// Fallback to string comparison for non-semver versions
	return currentVersion != latestVersion
}

// MeetsMinimumVersion checks if the current version meets the minimum required version
// Returns true if currentVersion >= minVersion
func (vc *VersionComparer) MeetsMinimumVersion(currentVersion, minVersion string) bool {
	// If no minimum version is specified, always pass
	if minVersion == "" {
		return true
	}

	// Handle empty current version
	if currentVersion == "" {
		return false
	}

	// Normalize versions
	currentVer := vc.normalizeVersion(currentVersion)
	minVer := vc.normalizeVersion(minVersion)

	// Try to parse as semantic versions
	current, err1 := semver.NewVersion(currentVer)
	minimum, err2 := semver.NewVersion(minVer)

	// If both versions are valid semver, use semantic comparison
	if err1 == nil && err2 == nil {
		return current.GreaterThan(minimum) || current.Equal(minimum)
	}

	// Fallback to string comparison for non-semver versions
	return currentVersion >= minVersion
}

// CompareVersions compares two versions and returns:
// -1 if version1 < version2
//
//	0 if version1 == version2
//	1 if version1 > version2
func (vc *VersionComparer) CompareVersions(version1, version2 string) int {
	// Handle empty versions
	if version1 == "" && version2 == "" {
		return 0
	}
	if version1 == "" {
		return -1
	}
	if version2 == "" {
		return 1
	}

	// Normalize versions
	ver1 := vc.normalizeVersion(version1)
	ver2 := vc.normalizeVersion(version2)

	// Try to parse as semantic versions
	v1, err1 := semver.NewVersion(ver1)
	v2, err2 := semver.NewVersion(ver2)

	// If both versions are valid semver, use semantic comparison
	if err1 == nil && err2 == nil {
		return v1.Compare(v2)
	}

	// Fallback to string comparison
	if version1 < version2 {
		return -1
	} else if version1 > version2 {
		return 1
	}
	return 0
}

// IsValidSemVer checks if a version string is a valid semantic version
func (vc *VersionComparer) IsValidSemVer(version string) bool {
	if version == "" {
		return false
	}

	normalized := vc.normalizeVersion(version)
	_, err := semver.NewVersion(normalized)
	return err == nil
}

// normalizeVersion adds 'v' prefix if missing and handles common version formats
func (vc *VersionComparer) normalizeVersion(version string) string {
	if version == "" {
		return version
	}

	// Remove any whitespace
	version = strings.TrimSpace(version)

	// Add 'v' prefix if missing (semver library expects it)
	if !strings.HasPrefix(version, "v") && !strings.HasPrefix(version, "V") {
		// Check if it looks like a version number (starts with digit)
		if len(version) > 0 && (version[0] >= '0' && version[0] <= '9') {
			version = "v" + version
		}
	}

	return version
}

// GetVersionInfo returns detailed information about a version
func (vc *VersionComparer) GetVersionInfo(version string) map[string]interface{} {
	info := map[string]interface{}{
		"original":   version,
		"normalized": vc.normalizeVersion(version),
		"is_semver":  false,
		"major":      nil,
		"minor":      nil,
		"patch":      nil,
		"prerelease": nil,
		"metadata":   nil,
	}

	if vc.IsValidSemVer(version) {
		info["is_semver"] = true
		if v, err := semver.NewVersion(vc.normalizeVersion(version)); err == nil {
			info["major"] = v.Major()
			info["minor"] = v.Minor()
			info["patch"] = v.Patch()
			info["prerelease"] = v.Prerelease()
			info["metadata"] = v.Metadata()
		}
	}

	return info
}

// SortVersions sorts a slice of version strings in ascending order
func (vc *VersionComparer) SortVersions(versions []string) []string {
	if len(versions) <= 1 {
		return versions
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]string, len(versions))
	copy(sorted, versions)

	// Simple bubble sort with semantic version comparison
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if vc.CompareVersions(sorted[j], sorted[j+1]) > 0 {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

