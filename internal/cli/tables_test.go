package cli

import (
	"bytes"
	"os"
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
