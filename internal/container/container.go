package container

import (
	"fmt"
	"os"
	"path/filepath"
)

// NeedsBuild determines if a container needs to be built
func NeedsBuild(name string, outputDir string) bool {
	tarFile := filepath.Join(outputDir, fmt.Sprintf("%s.tar.gz", name))
	_, err := os.Stat(tarFile)
	return os.IsNotExist(err)
}
