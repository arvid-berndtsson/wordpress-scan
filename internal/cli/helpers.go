package cli

import (
	"fmt"
	"os"
)

func ensureOutputDir(path string) error {
	if path == "" {
		return fmt.Errorf("output directory cannot be empty")
	}
	return os.MkdirAll(path, 0o755)
}
