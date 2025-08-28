package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Run-Panel/VerTree/internal/utils"
)

func main() {
	fmt.Println("🔄 VerTree Version Comparison Demo")
	fmt.Println("==================================")

	// 初始化版本比较器
	vc := utils.NewVersionComparer()

	// 测试版本比较功能
	fmt.Println("\n📊 Version Comparison Tests:")
	fmt.Println("----------------------------")

	testCases := []struct {
		current string
		latest  string
		desc    string
	}{
		{"1.0.0", "1.0.1", "Patch update"},
		{"1.0.0", "1.1.0", "Minor update"},
		{"1.0.0", "2.0.0", "Major update"},
		{"1.1.0", "1.0.0", "Downgrade (no update needed)"},
		{"2.0.0", "2.0.0", "Same version"},
		{"v1.0.0", "v1.0.1", "With v prefix"},
		{"1.0.0", "v1.0.1", "Mixed v prefix"},
		{"1.0.0-alpha", "1.0.0", "Prerelease to stable"},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", "Prerelease versions"},
		{"2.0.0+build.1", "2.0.0+build.2", "Build metadata"},
	}

	for i, tc := range testCases {
		needsUpdate := vc.IsUpdateNeeded(tc.current, tc.latest)
		comparison := vc.CompareVersions(tc.current, tc.latest)

		var compStr string
		switch comparison {
		case -1:
			compStr = "<"
		case 0:
			compStr = "="
		case 1:
			compStr = ">"
		}

		fmt.Printf("%2d. %s\n", i+1, tc.desc)
		fmt.Printf("    %s %s %s\n", tc.current, compStr, tc.latest)
		if needsUpdate {
			fmt.Printf("    ✅ Update needed: %s → %s\n", tc.current, tc.latest)
		} else {
			fmt.Printf("    ❌ No update needed\n")
		}
		fmt.Println()
	}

	// 测试版本排序
	fmt.Println("\n📋 Version Sorting Test:")
	fmt.Println("------------------------")

	versions := []string{
		"2.1.0",
		"1.0.0",
		"v1.2.0",
		"1.0.1",
		"2.0.0-beta.1",
		"2.0.0",
		"1.0.0-alpha",
		"v3.0.0",
	}

	fmt.Printf("Original order: %v\n", versions)

	sorted := vc.SortVersions(versions)
	fmt.Printf("Sorted order:   %v\n", sorted)

	// 测试最小版本要求
	fmt.Println("\n🔒 Minimum Version Requirements:")
	fmt.Println("--------------------------------")

	minVersionTests := []struct {
		current string
		minimum string
		desc    string
	}{
		{"1.2.0", "1.0.0", "Meets minimum requirement"},
		{"1.0.0", "1.2.0", "Below minimum requirement"},
		{"2.0.0", "1.5.0", "Exceeds minimum requirement"},
		{"1.0.0-beta", "1.0.0", "Prerelease vs stable"},
	}

	for i, tc := range minVersionTests {
		meets := vc.MeetsMinimumVersion(tc.current, tc.minimum)
		fmt.Printf("%d. %s\n", i+1, tc.desc)
		fmt.Printf("   Current: %s, Minimum: %s\n", tc.current, tc.minimum)
		if meets {
			fmt.Printf("   ✅ Meets requirement\n")
		} else {
			fmt.Printf("   ❌ Does not meet requirement\n")
		}
		fmt.Println()
	}

	// 测试版本信息提取
	fmt.Println("\n📝 Version Information Extraction:")
	fmt.Println("----------------------------------")

	infoTests := []string{
		"1.2.3",
		"v2.0.0-beta.1+build.123",
		"1.0.0-alpha.1",
		"3.1.4+metadata",
		"invalid-version",
	}

	for i, version := range infoTests {
		info := vc.GetVersionInfo(version)
		fmt.Printf("%d. Version: %s\n", i+1, version)
		fmt.Printf("   Original: %v\n", info["original"])
		fmt.Printf("   Normalized: %v\n", info["normalized"])
		fmt.Printf("   Is SemVer: %v\n", info["is_semver"])
		if info["is_semver"].(bool) {
			fmt.Printf("   Major: %v, Minor: %v, Patch: %v\n",
				info["major"], info["minor"], info["patch"])
			if info["prerelease"] != nil {
				fmt.Printf("   Prerelease: %v\n", info["prerelease"])
			}
			if info["metadata"] != nil {
				fmt.Printf("   Metadata: %v\n", info["metadata"])
			}
		}
		fmt.Println()
	}

	fmt.Println("✨ Demo completed successfully!")
}

func init() {
	// 确保能找到模块
	if _, err := os.Stat("../go.mod"); err != nil {
		// 尝试改变工作目录
		if err := os.Chdir(".."); err != nil {
			log.Fatal("Failed to find go.mod file")
		}
	}

	// 添加当前目录到 GOPATH
	wd, _ := os.Getwd()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		os.Setenv("GOPATH", wd)
	} else {
		os.Setenv("GOPATH", gopath+string(filepath.ListSeparator)+wd)
	}
}

