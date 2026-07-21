package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/go-github/v58/github"
	"github.com/revora/revora/internal/di"
	"github.com/spf13/cobra"
)

var releaseVersion string

func newReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Publish a draft release",
		RunE: func(cmd *cobra.Command, args []string) error {
			container := getContainer(cmd)
			if container == nil {
				return fmt.Errorf("container not initialised")
			}
			if container.GitHubClient == nil {
				return fmt.Errorf("GitHub client not available – run revora login")
			}
			return publishRelease(cmd.Context(), container)
		},
	}
	cmd.Flags().StringVarP(&releaseVersion, "version", "v", "", "Version to publish (default: latest draft)")
	return cmd
}

func publishRelease(ctx context.Context, container *di.Container) error {
	owner, repo, err := getOwnerRepo(container.Config.ProjectDir)
	if err != nil {
		return err
	}
	client := container.GitHubClient

	var targetRelease *github.RepositoryRelease
	if releaseVersion != "" {
		releases, _, err := client.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 100})
		if err != nil {
			return err
		}
		for _, rel := range releases {
			if rel.GetTagName() == releaseVersion {
				targetRelease = rel
				break
			}
		}
		if targetRelease == nil {
			return fmt.Errorf("release %s not found", releaseVersion)
		}
	} else {
		// find latest draft
		releases, _, err := client.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 10})
		if err != nil {
			return err
		}
		for _, rel := range releases {
			if rel.GetDraft() {
				targetRelease = rel
				break
			}
		}
		if targetRelease == nil {
			return fmt.Errorf("no draft releases found")
		}
	}

	tag := targetRelease.GetTagName()
	if tag == "" {
		return fmt.Errorf("release has no tag name")
	}

	// Ensure tag exists locally and on remote
	if err := ensureTag(container.Config.ProjectDir, tag); err != nil {
		return fmt.Errorf("ensure tag: %w", err)
	}

	// Publish
	targetRelease.Draft = github.Bool(false)
	if _, err := client.UpdateRelease(ctx, owner, repo, targetRelease.GetID(), targetRelease); err != nil {
		return err
	}
	fmt.Printf("Release %s published successfully.\n", tag)
	return nil
}

func ensureTag(projectDir, tag string) error {
	// Check if tag already exists
	out, _ := exec.Command("git", "-C", projectDir, "tag", "-l", tag).Output()
	if strings.TrimSpace(string(out)) == tag {
		// Tag exists; push to remote (no harm if already pushed)
		exec.Command("git", "-C", projectDir, "push", "origin", tag).Run()
		return nil
	}

	// Create tag at HEAD
	cmd := exec.Command("git", "-C", projectDir, "tag", tag)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create tag: %w", err)
	}

	// Push tag
	cmd = exec.Command("git", "-C", projectDir, "push", "origin", tag)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("push tag: %v\n%s", err, out)
	}
	return nil
}
