// cmd/revora-patch/main.go
package main

import (
	"archive/tar"
	"compress/gzip"
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
		os.Exit(1)
	}
}

func createCmd() *cobra.Command {
	var oldDir, newDir, patchFile string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a binary patch",
		RunE: func(cmd *cobra.Command, args []string) error {
			if oldDir == "" || newDir == "" || patchFile == "" {
				return fmt.Errorf("--old, --new, and --patch are required")
			}
			return createPatch(oldDir, newDir, patchFile)
		},
	}
	cmd.Flags().StringVar(&oldDir, "old", "", "old directory")
	cmd.Flags().StringVar(&newDir, "new", "", "new directory")
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
	f, err := os.Create(patchFile)
	if err != nil {
		return err
	}
	defer f.Close()
	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	base := filepath.Clean(newDir)
	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		// Determine if the file is new or changed relative to oldDir
		oldPath := filepath.Join(oldDir, rel)
		oldInfo, _ := os.Stat(oldPath)
		if info.IsDir() {
			return nil
		}
		if oldInfo != nil && !oldInfo.ModTime().Before(info.ModTime()) {
			// unchanged
			return nil
		}
		// Add to tar
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
	})
	return err
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
		defer outFile.Close()
		_, err = io.Copy(outFile, tr)
		if err != nil {
			return err
		}
	}
	return nil
}
