package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenFolder(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {

	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin": // MacOS
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("this platform is not supported")

	}

	return cmd.Start()
}
