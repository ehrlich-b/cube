package cli

import (
	"fmt"
	"time"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var solveCmd = newSolveCommand()

func newSolveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "solve [scramble]",
		Short: "Solve a 3x3 cube with the beginner method",
		Long: `Solve a physical 3x3 using the complete beginner layer-by-layer method.
Use cube learn with the same input for checkpoints and instructions.
Use --goal first-layer to stop after the white cross and corners.
With --start, the optional scramble is applied after the saved CFEN state.
Use --headless for space-separated solution moves, or --cfen for final state.`,
		Args: cobra.MaximumNArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			goal, _ := cmd.Flags().GetString("goal")
			if goal == "first-layer" {
				return runFirstLayerSolve(cmd, args)
			}
			if goal != "full" {
				return fmt.Errorf("unknown goal %q; choose full or first-layer", goal)
			}
			algorithm, _ := cmd.Flags().GetString("algorithm")
			if _, err := cube.GetSolver(algorithm); err != nil {
				return err
			}
			if algorithm != "beginner" {
				return fmt.Errorf("%s full-cube solver is not implemented; use --algorithm beginner", algorithm)
			}
			c, err := firstLayerInput(cmd, args)
			if err != nil {
				return err
			}
			started := time.Now()
			result, err := (&cube.BeginnerSolver{}).Solve(c)
			if err != nil {
				return err
			}
			c.ApplyMoves(result.Solution)
			if !c.IsSolved() {
				return fmt.Errorf("returned solution failed full-cube verification")
			}
			out := cmd.OutOrStdout()
			stateOnly, _ := cmd.Flags().GetBool("cfen")
			headless, _ := cmd.Flags().GetBool("headless")
			if stateOnly {
				state, err := cfen.GenerateCFEN(c)
				if err != nil {
					return err
				}
				fmt.Fprint(out, state)
			} else if headless {
				fmt.Fprint(out, cube.FormatMoves(result.Solution))
			} else {
				scramble := ""
				if len(args) > 0 {
					scramble = args[0]
				}
				fmt.Fprintln(out, "Solving 3x3x3 cube with scramble:", scramble)
				fmt.Fprintln(out, "Using algorithm: beginner")
				fmt.Fprintln(out, "Solution:", cube.FormatMoves(result.Solution))
				fmt.Fprintln(out, "Steps:", result.Steps)
				fmt.Fprintln(out, "Time:", time.Since(started))
				color, _ := cmd.Flags().GetBool("color")
				letters, _ := cmd.Flags().GetBool("letters")
				printLessonState(out, c, color, color && !letters, "full")
				fmt.Fprintln(out, "Use cube learn with the same input for the complete lesson.")
			}
			return nil
		},
	}
	cmd.Flags().String("goal", "full", "Solving goal: full cube or first-layer (3x3 beginner)")
	cmd.Flags().StringP("algorithm", "a", "beginner", "Solver: beginner (CFOP and Kociemba are not implemented)")
	cmd.Flags().IntP("dimension", "d", 3, "Cube dimension (beginner solver supports 3 only)")
	cmd.Flags().BoolP("color", "c", false, "Use colored output (Unicode blocks by default)")
	cmd.Flags().Bool("letters", false, "Use letters instead of Unicode blocks with --color")
	cmd.Flags().Bool("headless", false, "Output only solution moves")
	cmd.Flags().Bool("cfen", false, "Output final cube state as CFEN")
	cmd.Flags().String("start", "", "Starting cube state in concrete YB CFEN storage order")
	return cmd
}
