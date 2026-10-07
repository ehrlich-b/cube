package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var lessonMoveToken = regexp.MustCompile(`^(?:[URFDLB]w?|[MESxyz])(?:'|2)?$`)
var lessonFaceToken = regexp.MustCompile(`^(?:[WYROGB][1-9]?)+$`)

// Bound input before handing it to the general NxN parsers. In particular, a
// numeric CFEN run may otherwise expand to an arbitrarily large allocation.
func parseLessonMoves(text string) ([]cube.Move, error) {
	if len(text) > 8192 {
		return nil, fmt.Errorf("move input is too long (maximum 8192 characters)")
	}
	for _, token := range strings.Fields(text) {
		if !lessonMoveToken.MatchString(token) {
			return nil, fmt.Errorf("invalid 3x3 move %q; use R U F L D B, slice M E S, wide Rw, or rotations x y z, with ' or 2", token)
		}
	}
	return cube.ParseMoves(text)
}

func parseLessonState(text string) (*cube.Cube, error) {
	if len(text) > 128 || !strings.HasPrefix(text, "YB|") {
		return nil, fmt.Errorf("use a concrete 3x3 CFEN in YB storage order: U/R/F/D/L/B (as emitted by cube twist --cfen); rotations are read from the actual centers")
	}
	faces := strings.Split(strings.TrimPrefix(text, "YB|"), "/")
	if len(faces) != 6 {
		return nil, fmt.Errorf("CFEN needs six faces in U/R/F/D/L/B order")
	}
	for _, face := range faces {
		if !lessonFaceToken.MatchString(face) {
			return nil, fmt.Errorf("CFEN must contain real colors W Y R O G B and single-digit runs 1–9; wildcards are patterns, not physical states")
		}
	}
	state, err := cfen.ParseCFEN(text)
	if err != nil {
		return nil, err
	}
	if state.Dimension != 3 {
		return nil, fmt.Errorf("beginner lessons require a 3x3 cube")
	}
	c, err := state.ToCube()
	if err != nil {
		return nil, err
	}
	if err := cube.Validate3x3(c); err != nil {
		return nil, err
	}
	return c, nil
}

func firstLayerInput(cmd *cobra.Command, args []string) (*cube.Cube, error) {
	dimension, _ := cmd.Flags().GetInt("dimension")
	if dimension != 3 {
		return nil, fmt.Errorf("beginner lessons require --dimension 3")
	}
	start, _ := cmd.Flags().GetString("start")
	c := cube.NewCube(3)
	if start != "" {
		var err error
		c, err = parseLessonState(start)
		if err != nil {
			return nil, fmt.Errorf("starting state: %w", err)
		}
	}
	if len(args) > 0 {
		moves, err := parseLessonMoves(args[0])
		if err != nil {
			return nil, err
		}
		if err := c.ApplyMoves(moves); err != nil {
			return nil, err
		}
	}
	return c, nil
}

type lessonView struct {
	Steps []cube.FirstLayerStep
	Final *cube.Cube
	Moves []cube.Move
}

func planLesson(c *cube.Cube, goal string) (*lessonView, error) {
	if goal == "first-layer" {
		lesson, err := cube.PlanFirstLayer(c)
		if err != nil {
			return nil, err
		}
		return &lessonView{lesson.Steps, lesson.Final, lesson.Moves()}, nil
	}
	lesson, err := cube.PlanBeginner(c)
	if err != nil {
		return nil, err
	}
	return &lessonView{lesson.Steps, lesson.Final, lesson.Moves()}, nil
}

func newLearnCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "learn [scramble]", Short: "Learn a complete beginner solve on a 3x3 cube",
		Long: `Solve a 3x3 from your current cube state, one beginner checkpoint at a time:
white cross, white corners, middle edges, yellow cross, matching yellow edges,
placing yellow corners, and turning them upright. Use --goal first-layer for
just the white layer. The default full goal finishes all six faces.

To create a scramble, begin with a solved cube held white down, blue front.
A rotation in your scramble is supported. With --start, keep the orientation
represented by that state until the lesson instructs you to rotate.
Use --start with the CFEN emitted by cube twist --cfen to resume a physical state.
The optional scramble is applied AFTER --start.

Without --interactive, print the whole lesson. With --interactive, follow one
piece checkpoint at a time and record actual moves, undo, reset or save a CFEN.
Examples:
  cube learn "R U F2 L' B"
  cube learn "R U F2 L' B" --interactive --color
  cube learn --start "YB|Y9/R9/B9/W9/O9/G9"`,
		Args: cobra.MaximumNArgs(1), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			goal, _ := cmd.Flags().GetString("goal")
			if goal != "full" && goal != "first-layer" {
				return fmt.Errorf("unknown goal %q; choose full or first-layer", goal)
			}
			c, err := firstLayerInput(cmd, args)
			if err != nil {
				return err
			}
			interactive, _ := cmd.Flags().GetBool("interactive")
			color, _ := cmd.Flags().GetBool("color")
			letters, _ := cmd.Flags().GetBool("letters")
			out := cmd.OutOrStdout()
			printLessonIntro(out, goal)
			if interactive {
				return runLessonSession(cmd.InOrStdin(), out, c, color, color && !letters, goal)
			}
			lesson, err := planLesson(c, goal)
			if err != nil {
				return err
			}
			fmt.Fprintln(out, "Starting cube (net: U above; L F R B across; D below):")
			fmt.Fprintln(out, c.UnfoldedString(color, color && !letters))
			for i, step := range lesson.Steps {
				fmt.Fprintf(out, "Checkpoint %d: ", i+1)
				printLessonStep(out, step)
				state, _ := cfen.GenerateCFEN(step.After)
				fmt.Fprintf(out, "After this checkpoint: %s\n\n", state)
			}
			label := "Solution:"
			if goal == "first-layer" {
				label = "First-layer moves:"
			}
			fmt.Fprintln(out, label, cube.FormatMoves(lesson.Moves))
			printLessonState(out, lesson.Final, color, color && !letters, goal)
			return nil
		},
	}
	cmd.Flags().String("goal", "full", "Lesson goal: full cube or first-layer")
	cmd.Flags().IntP("dimension", "d", 3, "Cube dimension (beginner lessons support 3 only)")
	cmd.Flags().String("start", "", "Concrete CFEN in YB storage order; resume from a saved state")
	cmd.Flags().Bool("interactive", false, "Follow checkpoints with next, moves, undo, reset and state")
	cmd.Flags().BoolP("color", "c", false, "Use colored Unicode blocks")
	cmd.Flags().Bool("letters", false, "Use colored letters with --color")
	return cmd
}

func printLessonIntro(out io.Writer, goal string) {
	if goal == "first-layer" {
		fmt.Fprintln(out, "White first-layer lesson: cross + four corners on a 3x3.")
	} else {
		fmt.Fprintln(out, "Complete beginner 3x3 lesson: white layer, middle layer, yellow cross, yellow corners.")
	}
	fmt.Fprintln(out, "Start in the orientation represented by the input state. Rotate only when instructed.")
	fmt.Fprintln(out, "Target orientation: white down, yellow up, blue front, red right.")
	fmt.Fprintln(out, "R/L/U/D/F/B turn the right/left/top/bottom/front/back face clockwise as viewed directly at that face.")
	fmt.Fprintln(out, "' means counterclockwise, 2 means a half turn. x/y/z rotate the WHOLE cube like R/U/F; y brings the right side to the front.")
	fmt.Fprintln(out, "Keep holding the new orientation until instructed to return. Finish each move group before checking.")
	fmt.Fprintln(out)
}

func printLessonStep(out io.Writer, step cube.FirstLayerStep) {
	fmt.Fprintln(out, step.Title)
	for _, action := range step.Actions {
		if len(action.Moves) > 0 {
			fmt.Fprintf(out, "  %s\n  Do: %s\n", action.Instruction, cube.FormatMoves(action.Moves))
		}
	}
	fmt.Fprintln(out, "Check:", step.Check)
}

func printLessonState(out io.Writer, c *cube.Cube, color, unicode bool, goal string) {
	fmt.Fprintln(out, c.UnfoldedString(color, unicode))
	state, _ := cfen.GenerateCFEN(c)
	fmt.Fprintln(out, "Saved state:", state)
	if c.IsSolved() && goal == "full" {
		fmt.Fprintln(out, "Cube complete: all six faces are uniform and match their centers.")
		return
	}
	if cube.FirstLayerSolved(c) {
		if c.Faces[cube.Down][1][1] == cube.White {
			fmt.Fprintln(out, "First layer complete: white face and all four side bottom rows match their centers.")
		} else {
			fmt.Fprintln(out, "The white layer pieces are already solved. Follow any orientation checkpoint to put white down.")
		}
		if c.IsSolved() {
			fmt.Fprintln(out, "The whole cube is also solved.")
		} else if goal == "first-layer" {
			fmt.Fprintln(out, "Middle and last layers still need solving; this lesson stops at the first layer.")
		} else if cube.TwoLayersSolved(c) {
			fmt.Fprintln(out, "The first two layers are complete. Continue with the yellow layer.")
		} else {
			fmt.Fprintln(out, "Continue with the middle layer, then the yellow layer.")
		}
	}
}

func runLessonSession(in io.Reader, out io.Writer, c *cube.Cube, color, unicode bool, goal string) error {
	session, err := cube.NewFirstLayerSession(c)
	if goal == "full" {
		session, err = cube.NewBeginnerSession(c)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "Commands: next, hint, show, state, moves <actual moves>, undo, reset, help, quit.")
	fmt.Fprintln(out, "Perform the displayed moves on your cube, then type next to confirm them.")
	fmt.Fprintln(out, "If you made different moves, use moves <actual moves> instead of next; record everything since the current displayed state.")
	show := func() error {
		state := session.State()
		printLessonState(out, state, color, unicode, goal)
		lesson, err := planLesson(state, goal)
		if err == nil && len(lesson.Steps) > 0 {
			fmt.Fprint(out, "Next checkpoint: ")
			printLessonStep(out, lesson.Steps[0])
		}
		return err
	}
	if err := show(); err != nil {
		return err
	}
	reader := bufio.NewReaderSize(in, 1024)
	for {
		fmt.Fprint(out, "learn> ")
		input, err := readLessonLine(reader)
		if errors.Is(err, errLessonLineTooLong) {
			fmt.Fprintln(out, "No state changed: input line too long (maximum 8192 characters).")
			continue
		}
		if errors.Is(err, io.EOF) {
			state, _ := cfen.GenerateCFEN(session.State())
			fmt.Fprintln(out, "Input ended. Resume: cube learn --start '"+state+"' --goal "+goal+" --interactive")
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading lesson input: %w", err)
		}
		line := strings.TrimSpace(input)
		switch line {
		case "quit", "exit":
			state, _ := cfen.GenerateCFEN(session.State())
			fmt.Fprintln(out, "Resume: cube learn --start '"+state+"' --goal "+goal+" --interactive")
			return nil
		case "help":
			fmt.Fprintln(out, "next confirms the displayed group; moves <sequence> records what you actually did instead.")
			fmt.Fprintln(out, "hint/show display the state and next group; state saves CFEN; undo/reset print the physical inverse moves; quit saves a resume command.")
			continue
		case "state":
			state, _ := cfen.GenerateCFEN(session.State())
			fmt.Fprintln(out, state)
			continue
		case "hint", "show":
		case "next":
			step, err := session.Next()
			if err != nil {
				return err
			}
			if step != nil {
				fmt.Fprintln(out, "Confirmed:", step.Title)
			} else {
				if goal == "full" {
					fmt.Fprintln(out, "Cube already complete; no moves applied.")
				} else {
					fmt.Fprintln(out, "First layer already complete; no moves applied.")
				}
			}
		case "undo":
			moves, ok := session.Undo()
			if ok {
				fmt.Fprintln(out, "Undo on your physical cube:", cube.FormatMoves(moves))
			} else {
				fmt.Fprintln(out, "Nothing to undo.")
			}
		case "reset":
			moves := session.Reset()
			fmt.Fprintln(out, "Reset on your physical cube:", cube.FormatMoves(moves))
		default:
			if !strings.HasPrefix(line, "moves ") {
				fmt.Fprintln(out, "Unknown command. Type help; no state changed.")
				continue
			}
			moves, err := parseLessonMoves(strings.TrimSpace(strings.TrimPrefix(line, "moves ")))
			if err != nil {
				fmt.Fprintln(out, "No state changed:", err)
				continue
			}
			if err := session.Record(moves); err != nil {
				fmt.Fprintln(out, "No state changed:", err)
				continue
			}
			fmt.Fprintln(out, "Recorded actual moves; replanning from this state.")
		}
		if err := show(); err != nil {
			return err
		}
	}
}

var errLessonLineTooLong = errors.New("lesson input line too long")

// Consume an oversized line in bounded fragments, then allow the next command
// to recover. ReadString/ReadBytes would allocate the entire untrusted line.
func readLessonLine(reader *bufio.Reader) (string, error) {
	var line strings.Builder
	tooLong := false
	for {
		fragment, more, err := reader.ReadLine()
		if err != nil {
			return "", err
		}
		if line.Len()+len(fragment) > 8192 {
			tooLong = true
		}
		if !tooLong {
			line.Write(fragment)
		}
		if !more {
			if tooLong {
				return "", errLessonLineTooLong
			}
			return line.String(), nil
		}
	}
}

func runFirstLayerSolve(cmd *cobra.Command, args []string) error {
	algorithm, _ := cmd.Flags().GetString("algorithm")
	if !cmd.Flags().Changed("algorithm") {
		algorithm = "beginner"
	}
	if method := cmd.Flags().Lookup("method"); method != nil && cmd.Flags().Changed("method") {
		if cmd.Flags().Changed("algorithm") && algorithm != method.Value.String() {
			return fmt.Errorf("--method and --algorithm select different solvers")
		}
		algorithm = method.Value.String()
	}
	if algorithm != "beginner" {
		return fmt.Errorf("--goal first-layer requires --algorithm beginner")
	}
	c, err := firstLayerInput(cmd, args)
	if err != nil {
		return err
	}
	lesson, err := cube.PlanFirstLayer(c)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	stateOnly, _ := cmd.Flags().GetBool("cfen")
	headless, _ := cmd.Flags().GetBool("headless")
	switch {
	case stateOnly:
		state, _ := cfen.GenerateCFEN(lesson.Final)
		fmt.Fprint(out, state)
	case headless:
		fmt.Fprint(out, cube.FormatMoves(lesson.Moves()))
	default:
		fmt.Fprintln(out, "Goal: white first layer (3x3 beginner)")
		fmt.Fprintln(out, "First-layer moves:", cube.FormatMoves(lesson.Moves()))
		color, _ := cmd.Flags().GetBool("color")
		letters, _ := cmd.Flags().GetBool("letters")
		printLessonState(out, lesson.Final, color, color && !letters, "first-layer")
		fmt.Fprintln(out, "Use cube learn --goal first-layer with the same input for piece-by-piece instructions.")
	}
	return nil
}

func init() { rootCmd.AddCommand(newLearnCommand()) }
