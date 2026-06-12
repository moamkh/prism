package version

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetVersion() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "0.0.0"
	}
	// From rev_core/internal/version/ go up 3 dirs to repo root
	dir := filepath.Dir(filename)
	versionFile := filepath.Join(dir, "..", "..", "..", "VERSION")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "0.0.0"
	}
	return strings.TrimSpace(string(data))
}
