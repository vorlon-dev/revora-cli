package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func newVersionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "versions",
		Short: "List all releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			if container.GitHubClient == nil {
				return fmt.Errorf("GitHub client not available – run revora login first")
			}

			// We need owner/repo from git remote
			owner, repo, err := getOwnerRepo(container.Config.ProjectDir)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			releases, _, err := container.GitHubClient.ListReleases(ctx, owner, repo, nil)
			if err != nil {
				return fmt.Errorf("fetch releases: %w", err)
			}
			if len(releases) == 0 {
				fmt.Println("No releases found.")
				return nil
			}
			for _, r := range releases {
				fmt.Printf("%s\t(%s)\n", r.GetTagName(), r.GetPublishedAt().Format("2006-01-02"))
			}
			return nil
		},
	}
}

func getOwnerRepo(projectDir string) (string, string, error) {
	// Quick and dirty: get git remote origin URL and parse owner/repo
	cmd := exec.Command("git", "-C", projectDir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git remote: %w", err)
	}
	url := strings.TrimSpace(string(out))
	// parse "https://github.com/owner/repo.git" or "git@github.com:owner/repo.git"
	parts := strings.Split(url, "github.com/")
	if len(parts) < 2 {
		parts = strings.Split(url, "github.com:")
	}
	if len(parts) < 2 {
		return "", "", fmt.Errorf("cannot parse GitHub remote URL: %s", url)
	}
	repoPath := strings.TrimSuffix(parts[1], ".git")
	parts = strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo path: %s", repoPath)
	}
	return parts[0], parts[1], nil
}
