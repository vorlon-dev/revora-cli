package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Manager struct {
	ProjectDir  string
	GitHubToken string
	Platform    string
	SigningKey  string
	PatchEngine string // path to revora-patch binary (default "revora-patch")
}

func Load() (*Manager, error) {
	v := viper.New()

	start, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working dir: %w", err)
	}

	v.SetConfigName("revora")
	v.SetConfigType("yaml")

	// Add search paths from current dir up to filesystem root
	dir := start
	for {
		v.AddConfigPath(dir)
		parent := filepath.Dir(dir)
		if parent == dir { // reached root (e.g., "/" or "C:\")
			break
		}
		dir = parent
	}

	v.AddConfigPath("$HOME/.revora")
	v.AddConfigPath("/etc/revora")

	v.SetEnvPrefix("REVORA")
	v.AutomaticEnv()
	v.SetDefault("platform", "")
	v.SetDefault("signing_key", "")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	projectDir := start
	if configFile := v.ConfigFileUsed(); configFile != "" {
		projectDir = filepath.Dir(configFile)
	}

	token := os.Getenv("REVORA_GITHUB_TOKEN")
	patchEngine := os.Getenv("REVORA_PATCH_ENGINE")
	if patchEngine == "" {
		patchEngine = "revora-patch"
	}

	return &Manager{
		ProjectDir:  projectDir,
		GitHubToken: token,
		Platform:    v.GetString("platform"),
		SigningKey:  v.GetString("signing_key"),
		PatchEngine: patchEngine,
	}, nil
}
