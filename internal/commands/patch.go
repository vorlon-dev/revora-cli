// internal/commands/patch.go
package commands

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-github/v58/github"
	"github.com/revora/revora/internal/config"
	"github.com/revora/revora/internal/crypto"
	"github.com/revora/revora/internal/di"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newPatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "patch",
		Short: "Create and upload a patch release",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			return createPatch(cmd.Context(), container)
		},
	}
}

func createPatch(ctx context.Context, container *di.Container) error {
	logger := container.Logger
	cfg := container.Config

	// 1. Determine old and new build directories
	oldDir := filepath.Join(cfg.ProjectDir, ".revora", "cache", "previous")
	newDir := filepath.Join(cfg.ProjectDir, "build")

	logger.Info("Checking directories",
		zap.String("old", oldDir),
		zap.String("new", newDir),
	)

	// If the previous build cache doesn't exist, create it from the current build
	// This happens on the very first patch (no baseline)
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		logger.Warn("No previous build found. This appears to be your first patch.")
		fmt.Println("⚠️  First patch detected. Using current build as baseline (this patch may be large).")
		if err := os.MkdirAll(oldDir, 0755); err != nil {
			return fmt.Errorf("create previous cache dir: %w", err)
		}
		// Copy current build to previous
		if err := copyDir(newDir, oldDir); err != nil {
			return fmt.Errorf("copy current build to previous: %w", err)
		}
		fmt.Println("✅ Baseline created. Future patches will be small and fast.")
		// Since we just made old identical to new, there's nothing to patch.
		// We'll exit gracefully and let the user know.
		fmt.Println("No changes to patch (baseline established). Run 'revora patch' again after making code changes.")
		return nil
	}

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		return fmt.Errorf("current build not found at %s – run 'revora build' first", newDir)
	}

	// 2. Generate patch using the patch engine
	patchFile := filepath.Join(cfg.ProjectDir, ".revora", "update.patch")
	engine := cfg.PatchEngine
	if engine == "" {
		engine = "revora-patch"
	}

	logger.Info("Creating patch (engine will show progress)", zap.String("engine", engine))
	cmd := exec.Command(engine, "create", "--old", oldDir, "--new", newDir, "--patch", patchFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr // engine prints per-file progress to stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("patch creation failed: %w", err)
	}

	// 3. Compute SHA-256 of the patch file
	logger.Info("Computing patch hash...")
	hash, err := sha256File(patchFile)
	if err != nil {
		return err
	}
	logger.Info("Patch hash", zap.String("sha256", hash))

	// 4. Determine version numbers
	lastVersion := getLastVersion(cfg.ProjectDir)
	nextVersion := getNextVersion(cfg.ProjectDir)
	logger.Info("Versioning",
		zap.String("previous", lastVersion),
		zap.String("next", nextVersion),
	)

	// 5. Create manifest
	manifest := map[string]interface{}{
		"version":      nextVersion,
		"base_version": lastVersion,
		"platform":     cfg.Platform,
		"patch_sha256": hash,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	// 6. Sign manifest
	logger.Info("Signing manifest...")
	signature, err := signManifest(manifestBytes, cfg)
	if err != nil {
		return fmt.Errorf("sign manifest: %w", err)
	}

	// 7. Save manifest and signature locally
	manifestFile := filepath.Join(cfg.ProjectDir, ".revora", "manifest.json")
	sigFile := manifestFile + ".sig"
	if err := os.WriteFile(manifestFile, manifestBytes, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(sigFile, signature, 0644); err != nil {
		return err
	}

	// 8. Upload to GitHub (if a client is available)
	if container.GitHubClient == nil {
		return fmt.Errorf("GitHub client not available – run revora login")
	}
	owner, repo, err := getOwnerRepo(cfg.ProjectDir)
	if err != nil {
		return err
	}

	logger.Info("Creating draft release on GitHub...")
	release, err := container.GitHubClient.CreateRelease(ctx, owner, repo, &github.RepositoryRelease{
		TagName:    github.String(nextVersion),
		Name:       github.String(nextVersion),
		Draft:      github.Bool(true),
		Prerelease: github.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("create release: %w", err)
	}

	// Upload assets with size logging
	uploadFile := func(name string) error {
		f, err := os.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}
		logger.Info("Uploading",
			zap.String("file", filepath.Base(name)),
			zap.Int64("size_bytes", fi.Size()),
		)
		_, err = container.GitHubClient.UploadReleaseAsset(ctx, owner, repo, release.GetID(), &github.UploadOptions{
			Name: filepath.Base(name),
		}, f)
		return err
	}

	if err := uploadFile(patchFile); err != nil {
		return fmt.Errorf("upload patch: %w", err)
	}
	if err := uploadFile(manifestFile); err != nil {
		return fmt.Errorf("upload manifest: %w", err)
	}
	if err := uploadFile(sigFile); err != nil {
		return fmt.Errorf("upload signature: %w", err)
	}

	logger.Info("Draft release created", zap.String("version", nextVersion))
	fmt.Printf("Draft release %s created. Use 'revora release' to publish.\n", nextVersion)
	return nil
}

// --- helper functions ---

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func signManifest(data []byte, cfg *config.Manager) ([]byte, error) {
	privPath := filepath.Join(cfg.ProjectDir, ".revora", "keys", "private.pem")
	priv, err := crypto.LoadPrivateKey(privPath)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}
	sig, err := crypto.Sign(priv, data)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	return sig, nil
}

func getLastVersion(projectDir string) string {
	out, err := exec.Command("git", "-C", projectDir, "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		return "0.0.0"
	}
	return strings.TrimSpace(string(out))
}

func getNextVersion(projectDir string) string {
	last := getLastVersion(projectDir)
	if last == "" || last == "0.0.0" {
		return "0.0.1"
	}
	parts := strings.Split(last, ".")
	if len(parts) != 3 {
		return "0.0.1"
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "0.0.1"
	}
	parts[2] = strconv.Itoa(patch + 1)
	return strings.Join(parts, ".")
}

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
