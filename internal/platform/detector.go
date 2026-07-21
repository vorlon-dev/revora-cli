package platform

import (
	"os"
	"path/filepath"
)

type Detector interface {
	Detect(projectDir string) (string, error)
}

type detector struct{}

func NewDetector() Detector {
	return &detector{}
}

func (d *detector) Detect(projectDir string) (string, error) {
	// Check common project files in order
	checks := []struct {
		files    []string
		platform string
	}{
		{[]string{"build.gradle", "build.gradle.kts"}, "android"},
		{[]string{"pubspec.yaml"}, "flutter"},
		{[]string{"settings.gradle", "settings.gradle.kts"}, "android"}, // generic gradle
		{[]string{"package.json"}, "react-native"},
		{[]string{"build.zig"}, "unknown"},
	}

	for _, check := range checks {
		for _, filename := range check.files {
			if _, err := os.Stat(filepath.Join(projectDir, filename)); err == nil {
				return check.platform, nil
			}
		}
	}
	return "", &ErrUnknownPlatform{projectDir}
}

type ErrUnknownPlatform struct {
	Dir string
}

func (e *ErrUnknownPlatform) Error() string {
	return "could not detect platform in " + e.Dir
}
