package utils

import (
	"testing"
)

func TestVersionComparer_IsUpdateNeeded(t *testing.T) {
	vc := NewVersionComparer()

	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		expected       bool
	}{
		// Semantic version tests
		{"Patch update needed", "1.0.0", "1.0.1", true},
		{"Minor update needed", "1.0.0", "1.1.0", true},
		{"Major update needed", "1.0.0", "2.0.0", true},
		{"No update needed - same version", "1.0.0", "1.0.0", false},
		{"No update needed - current newer", "1.1.0", "1.0.0", false},

		// Version with 'v' prefix
		{"With v prefix - update needed", "v1.0.0", "v1.0.1", true},
		{"Mixed v prefix - update needed", "1.0.0", "v1.0.1", true},
		{"With v prefix - no update", "v1.0.1", "v1.0.0", false},

		// Prerelease versions
		{"Prerelease to stable", "1.0.0-alpha", "1.0.0", true},
		{"Prerelease to prerelease", "1.0.0-alpha.1", "1.0.0-alpha.2", true},
		{"Stable to prerelease", "1.0.0", "1.1.0-beta", true},

		// Non-semver fallback
		{"Non-semver different", "1.0", "1.1", true},
		{"Non-semver same", "1.0", "1.0", false},

		// Edge cases
		{"Empty current", "", "1.0.0", true},
		{"Empty latest", "1.0.0", "", true},
		{"Both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vc.IsUpdateNeeded(tt.currentVersion, tt.latestVersion)
			if result != tt.expected {
				t.Errorf("IsUpdateNeeded(%q, %q) = %v, want %v",
					tt.currentVersion, tt.latestVersion, result, tt.expected)
			}
		})
	}
}

func TestVersionComparer_MeetsMinimumVersion(t *testing.T) {
	vc := NewVersionComparer()

	tests := []struct {
		name           string
		currentVersion string
		minVersion     string
		expected       bool
	}{
		// Semantic version tests
		{"Meets minimum - same version", "1.0.0", "1.0.0", true},
		{"Meets minimum - higher patch", "1.0.1", "1.0.0", true},
		{"Meets minimum - higher minor", "1.1.0", "1.0.0", true},
		{"Meets minimum - higher major", "2.0.0", "1.0.0", true},
		{"Below minimum - patch", "1.0.0", "1.0.1", false},
		{"Below minimum - minor", "1.0.0", "1.1.0", false},
		{"Below minimum - major", "1.0.0", "2.0.0", false},

		// Version with 'v' prefix
		{"With v prefix - meets", "v1.1.0", "v1.0.0", true},
		{"Mixed v prefix - meets", "1.1.0", "v1.0.0", true},
		{"With v prefix - below", "v1.0.0", "v1.1.0", false},

		// Edge cases
		{"No minimum required", "1.0.0", "", true},
		{"Empty current with minimum", "", "1.0.0", false},
		{"Both empty", "", "", true},

		// Non-semver fallback
		{"Non-semver meets", "1.1", "1.0", true},
		{"Non-semver below", "1.0", "1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vc.MeetsMinimumVersion(tt.currentVersion, tt.minVersion)
			if result != tt.expected {
				t.Errorf("MeetsMinimumVersion(%q, %q) = %v, want %v",
					tt.currentVersion, tt.minVersion, result, tt.expected)
			}
		})
	}
}

func TestVersionComparer_CompareVersions(t *testing.T) {
	vc := NewVersionComparer()

	tests := []struct {
		name     string
		version1 string
		version2 string
		expected int
	}{
		// Semantic version tests
		{"Same versions", "1.0.0", "1.0.0", 0},
		{"First greater - patch", "1.0.1", "1.0.0", 1},
		{"First greater - minor", "1.1.0", "1.0.0", 1},
		{"First greater - major", "2.0.0", "1.0.0", 1},
		{"First smaller - patch", "1.0.0", "1.0.1", -1},
		{"First smaller - minor", "1.0.0", "1.1.0", -1},
		{"First smaller - major", "1.0.0", "2.0.0", -1},

		// Edge cases
		{"Empty versions", "", "", 0},
		{"First empty", "", "1.0.0", -1},
		{"Second empty", "1.0.0", "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vc.CompareVersions(tt.version1, tt.version2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %v, want %v",
					tt.version1, tt.version2, result, tt.expected)
			}
		})
	}
}

func TestVersionComparer_IsValidSemVer(t *testing.T) {
	vc := NewVersionComparer()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"Valid semver", "1.0.0", true},
		{"Valid with v prefix", "v1.0.0", true},
		{"Valid with prerelease", "1.0.0-alpha", true},
		{"Valid with metadata", "1.0.0+build.1", true},
		{"Valid complex", "1.0.0-alpha.1+build.1", true},

		{"Valid - missing patch (auto-completed)", "1.0", true},
		{"Invalid - non-numeric", "1.0.a", false},
		{"Invalid - empty", "", false},
		{"Invalid - just text", "latest", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vc.IsValidSemVer(tt.version)
			if result != tt.expected {
				t.Errorf("IsValidSemVer(%q) = %v, want %v",
					tt.version, result, tt.expected)
			}
		})
	}
}

func TestVersionComparer_SortVersions(t *testing.T) {
	vc := NewVersionComparer()

	tests := []struct {
		name     string
		versions []string
		expected []string
	}{
		{
			name:     "Semantic versions",
			versions: []string{"2.0.0", "1.0.0", "1.1.0", "1.0.1"},
			expected: []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"},
		},
		{
			name:     "With v prefix",
			versions: []string{"v2.0.0", "v1.0.0", "v1.1.0"},
			expected: []string{"v1.0.0", "v1.1.0", "v2.0.0"},
		},
		{
			name:     "Mixed versions",
			versions: []string{"2.0.0", "v1.0.0", "1.1.0"},
			expected: []string{"v1.0.0", "1.1.0", "2.0.0"},
		},
		{
			name:     "Single version",
			versions: []string{"1.0.0"},
			expected: []string{"1.0.0"},
		},
		{
			name:     "Empty slice",
			versions: []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vc.SortVersions(tt.versions)
			if len(result) != len(tt.expected) {
				t.Errorf("SortVersions() returned slice of length %d, want %d",
					len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("SortVersions() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}
