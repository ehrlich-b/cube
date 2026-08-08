package cli

// commands_test.go guards the CLI command surface. Commands are registered from
// two places (root.go and each command file's own init()), which makes it easy
// to accidentally register the same command twice — that is exactly the bug this
// test exists to catch, so `cube --help` always lists each command once.

import "testing"

func TestNoDuplicateCommands(t *testing.T) {
	seen := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		name := c.Name()
		if seen[name] {
			t.Errorf("command %q is registered more than once (check root.go and %s.go init())", name, name)
		}
		seen[name] = true
	}
}
