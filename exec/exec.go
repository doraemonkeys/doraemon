package exec

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenUrl(uri string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "start", uri)
		return cmd.Start()
	case "darwin":
		cmd := exec.Command("open", uri)
		return cmd.Start()
	case "linux":
		cmd := exec.Command("xdg-open", uri)
		return cmd.Start()
	default:
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}
}
