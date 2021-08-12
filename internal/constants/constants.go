package constants

import (
	"os"
	"path/filepath"
)

var homeDir, _ = os.UserHomeDir()
var StacksDir = filepath.Join(homeDir, ".firefly", "stacks")
