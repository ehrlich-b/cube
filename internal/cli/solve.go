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
		Short: "Solve a 3x3 cube with Kociemba, beginner or CFOP",
		Long: `Solve a physical 3x3 using Kociemba's two-phase method (default).
Select --method beginner for the complete beginner layer-by-layer method.
Select --method cfop for verified Cross, F2L 1-4, OLL and PLL stages with case names.
Use cube learn with the same input for checkpoints and instructions.
Use --goal first-layer to stop after the white cross and corners.
With --start, the optional scramble is applied after the saved CFEN state.
Use --headless for space-separated solution moves, or --cfen for final state.`,
		Args: cobra.MaximumNArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			goal, _ := cmd.Flags().GetString("goal")
			optimal, _ := cmd.Flags().GetBool("optimal")
			if goal == "first-layer" {
				if optimal {
					return fmt.Errorf("--optimal requires --goal full")
				}
				return runFirstLayerSolve(cmd, args)
			}
			if goal != "full" {
				return fmt.Errorf("unknown goal %q; choose full or first-layer", goal)
			}
			algorithm, _ := cmd.Flags().GetString("algorithm")
			method, _ := cmd.Flags().GetString("method")
			if cmd.Flags().Changed("method") {
				if cmd.Flags().Changed("algorithm") && algorithm != method {
					return fmt.Errorf("--method and --algorithm select different solvers")
				}
				algorithm = method
			}
			solver, err := cube.GetSolver(algorithm)
			if err != nil {
				return err
			}
			if optimal && (cmd.Flags().Changed("algorithm") || cmd.Flags().Changed("method")) && algorithm != "kociemba" {
				return fmt.Errorf("--optimal cannot be combined with --method %s", algorithm)
			}
			c, err := firstLayerInput(cmd, args)
			if err != nil {
				return err
			}
			started := time.Now()
			var result *cube.SolverResult
			if optimal {
				limit, _ := cmd.Flags().GetDuration("time-limit")
				result, err = cube.SolveOptimal(c, limit)
				algorithm = "optimal"
			} else {
				result, err = solver.Solve(c)
			}
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
				fmt.Fprintln(out, "Using algorithm:", algorithm)
				fmt.Fprintln(out, "Solution:", cube.FormatMoves(result.Solution))
				for _, stage := range result.Stages {
					fmt.Fprintf(out, "%s: %s (%d turns)\n  %s\n", stage.Name, stage.CaseName(), cube.TurnCount(stage.Moves), cube.FormatMoves(stage.Moves))
				}
				fmt.Fprintln(out, "Steps:", result.Steps)
				fmt.Fprintln(out, "Time:", time.Since(started))
				color, _ := cmd.Flags().GetBool("color")
				letters, _ := cmd.Flags().GetBool("letters")
				printLessonState(out, c, color, color && !letters, "full")
				fmt.Fprintln(out, "Use cube learn with the same input for the complete beginner lesson.")
			}
			return nil
		},
	}
	cmd.Flags().String("goal", "full", "Solving goal: full cube or first-layer (3x3 beginner)")
	cmd.Flags().StringP("algorithm", "a", "kociemba", "Solver: kociemba, beginner or cfop")
	cmd.Flags().String("method", "kociemba", "Solving method (alias for --algorithm): kociemba, beginner or cfop")
	cmd.Flags().Bool("optimal", false, "Prove a shortest 3x3 solution with IDA* (errors if the time limit expires)")
	cmd.Flags().Duration("time-limit", time.Second, "Time limit for --optimal, including table initialization")
	cmd.Flags().IntP("dimension", "d", 3, "Cube dimension (full solving supports 3 only)")
	cmd.Flags().BoolP("color", "c", false, "Use colored output (Unicode blocks by default)")
	cmd.Flags().Bool("letters", false, "Use letters instead of Unicode blocks with --color")
	cmd.Flags().Bool("headless", false, "Output only solution moves")
	cmd.Flags().Bool("cfen", false, "Output final cube state as CFEN")
	cmd.Flags().String("start", "", "Starting cube state in concrete YB CFEN storage order")
	return cmd
}
