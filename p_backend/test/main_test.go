package test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	ensureBackendRoot()
	os.Exit(m.Run())
}

func ensureBackendRoot() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			_ = os.Chdir(wd)
			return
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return
		}
		wd = parent
	}
}
