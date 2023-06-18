package earthdir

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/earthly/earthly/util/fileutil"
)

const (
	earthfileName = "Earthfile"
)

func InOrAbove(dir string, upToDir string, closest bool) (string, error) {
	if !strings.HasPrefix(dir, upToDir) {
		return "", fmt.Errorf("dir %s is not in %s", dir, upToDir)
	}

	lastFoundDir := ""
	for d := dir; d != upToDir; d = filepath.Dir(d) {
		ep := filepath.Join(d, earthfileName)
		exists, err := fileutil.FileExists(ep)
		if err != nil {
			return "", err
		}
		if exists {
			if closest {
				return d, nil
			}

			lastFoundDir = d
		}
	}

	if lastFoundDir == "" {
		return "", fmt.Errorf("no Earthfile found in or above %s", dir)
	}

	return lastFoundDir, nil
}
