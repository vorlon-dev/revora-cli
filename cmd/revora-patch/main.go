// cmd/revora-patch/main.go
package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "revora-patch"}
	root.AddCommand(createCmd())
	root.AddCommand(applyCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func createCmd() *cobra.Command {
	var oldDir, newDir, patchFile string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a binary patch (diff between two directories)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if oldDir == "" || newDir == "" || patchFile == "" {
				return fmt.Errorf("--old, --new, and --patch are required")
			}
			return createPatch(oldDir, newDir, patchFile)
		},
	}
	cmd.Flags().StringVar(&oldDir, "old", "", "old directory (baseline)")
	cmd.Flags().StringVar(&newDir, "new", "", "new directory (current)")
	cmd.Flags().StringVar(&patchFile, "patch", "", "output patch file")
	return cmd
}

func applyCmd() *cobra.Command {
	var oldDir, newDir, patchFile string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a binary patch",
		RunE: func(cmd *cobra.Command, args []string) error {
			if oldDir == "" || newDir == "" || patchFile == "" {
				return fmt.Errorf("--old, --new, and --patch are required")
			}
			return applyPatch(oldDir, newDir, patchFile)
		},
	}
	cmd.Flags().StringVar(&oldDir, "old", "", "old directory")
	cmd.Flags().StringVar(&newDir, "new", "", "new directory")
	cmd.Flags().StringVar(&patchFile, "patch", "", "input patch file")
	return cmd
}

func createPatch(oldDir, newDir, patchFile string) error {
	out, err := os.Create(patchFile)
	if err != nil {
		return err
	}
	defer out.Close()
	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	base := filepath.Clean(newDir)
	var countAdded, countTotal int

	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "walk error: %v\n", err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		countTotal++
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}

		oldPath := filepath.Join(oldDir, rel)
		changed, err := fileChanged(oldPath, path)
		if err != nil || changed {
			// file is new or changed – add to patch
			countAdded++
			fmt.Fprintf(os.Stderr, "[%d/%d] adding  %s\n", countAdded, countTotal, rel)

			header, err := tar.FileInfoHeader(info, rel)
			if err != nil {
				return err
			}
			header.Name = rel
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			src, err := os.Open(path)
			if err != nil {
				return err
			}
			defer src.Close()
			_, err = io.Copy(tw, src)
			return err
		}
		fmt.Fprintf(os.Stderr, "[%d/%d] unchanged  %s\n", countAdded, countTotal, rel)
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Patch complete: %d files changed/new out of %d total.\n", countAdded, countTotal)
	return nil
}

func fileChanged(oldPath, newPath string) (bool, error) {
	oldHash, err := fileSHA256(oldPath)
	if err != nil {
		// old file doesn't exist or can't be read → treat as changed
		return true, nil
	}
	newHash, err := fileSHA256(newPath)
	if err != nil {
		return false, err
	}
	return oldHash != newHash, nil
}

func fileSHA256(path string) (string, error) {
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

func applyPatch(oldDir, newDir, patchFile string) error {
	f, err := os.Open(patchFile)
	if err != nil {
		return err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(newDir, header.Name)
		os.MkdirAll(filepath.Dir(target), 0755)
		if header.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		outFile, err := os.Create(target)
		if err != nil {
			return err
		}
		_, err = io.Copy(outFile, tr)
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
