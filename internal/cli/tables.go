package cli

import (
	"fmt"

	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

func newTablesCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "tables", Short: "Manage optional solver tables"}
	build := &cobra.Command{
		Use: "build", Short: "Build and persist the optional large phase-one table", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			large, _ := cmd.Flags().GetBool("large")
			if !large {
				return fmt.Errorf("compact tables ship with cube; use --large to opt into the 140.67 MiB table")
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "Building optional large tables: 140.67 MiB of cache; allow a few minutes and about 1.6 GiB of temporary memory.")
			bytes, err := cube.BuildLargePhase1Tables()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Large phase-one table ready: %d bytes. Set CUBE_LARGE_TABLES=1 to use it.\n", bytes)
			return nil
		},
	}
	build.Flags().Bool("large", false, "Explicitly build the optional 140.67 MiB phase-one cache")
	cmd.AddCommand(build)
	return cmd
}

func init() { rootCmd.AddCommand(newTablesCommand()) }
