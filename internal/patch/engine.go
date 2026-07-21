package patch

import (
"fmt"
"os/exec"
)

type Engine interface {
Apply(oldDir, newDir, patchFile string) error
Create(oldDir, newDir, patchFile string) error
}

type localEngine struct {
binaryPath string
}

func NewLocalEngine(path string) Engine {
return &localEngine{binaryPath: path}
}

func (e *localEngine) Apply(oldDir, newDir, patchFile string) error {
cmd := exec.Command(e.binaryPath, "apply", "--old", oldDir, "--new", newDir, "--patch", patchFile)
out, err := cmd.CombinedOutput()
if err != nil {
return fmt.Errorf("patch apply failed: %v\n%s", err, out)
}
return nil
}

func (e *localEngine) Create(oldDir, newDir, patchFile string) error {
cmd := exec.Command(e.binaryPath, "create", "--old", oldDir, "--new", newDir, "--patch", patchFile)
out, err := cmd.CombinedOutput()
if err != nil {
return fmt.Errorf("patch creation failed: %v\n%s", err, out)
}
return nil
}
