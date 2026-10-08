package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTablesBuildRequiresExplicitLargeFlag(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("CUBE_CACHE_DIR", cache)
	cmd := newTablesCommand()
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"build"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--large") {
		t.Fatal("large build must require an explicit flag", err)
	}
	files, err := os.ReadDir(cache)
	if err != nil || len(files) != 0 {
		t.Fatal("unapproved table generation", files, err)
	}
}

func TestTablesBuildReportsUnwritableCache(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(cache, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CUBE_CACHE_DIR", cache)
	t.Setenv("CUBE_SMALL_TABLES", "0")
	cmd := newTablesCommand()
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	var out, diagnostic bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&diagnostic)
	cmd.SetArgs([]string{"build", "--large"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "persist large phase-one table") {
		t.Fatal("large build discarded cache failure", err)
	}
	if strings.Contains(out.String(), "ready") {
		t.Fatal("failed build reported ready", out.String())
	}
	if data, err := os.ReadFile(cache); err != nil || string(data) != "keep" {
		t.Fatal("failed build changed existing file", err)
	}
}
