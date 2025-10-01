package cube

import (
	"testing"
)

// BenchmarkBeginnerSolver benchmarks the beginner solver on various scramble complexities
func BenchmarkBeginnerSolver(b *testing.B) {
	benchmarks := []struct {
		name     string
		scramble string
	}{
		{"1move", "R"},
		{"2moves", "R U"},
		{"3moves", "R U F"},
		{"4moves", "R U R' U'"},
		{"5moves", "R U R' U' F"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			moves, _ := ParseScramble(bm.scramble)
			solver := &BeginnerSolver{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cube := NewCube(3)
				cube.ApplyMoves(moves)
				_, err := solver.Solve(cube)
				if err != nil {
					b.Fatalf("Solver failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkCFOPSolver benchmarks the CFOP solver on various scramble complexities
func BenchmarkCFOPSolver(b *testing.B) {
	benchmarks := []struct {
		name     string
		scramble string
	}{
		{"1move", "F"},
		{"2moves", "R U"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			moves, _ := ParseScramble(bm.scramble)
			solver := &CFOPSolver{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cube := NewCube(3)
				cube.ApplyMoves(moves)
				_, err := solver.Solve(cube)
				if err != nil {
					b.Fatalf("Solver failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkKociembaSolver benchmarks the Kociemba solver on various scramble complexities
func BenchmarkKociembaSolver(b *testing.B) {
	benchmarks := []struct {
		name     string
		scramble string
	}{
		{"1move", "R"},
		{"2moves", "R U"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			moves, _ := ParseScramble(bm.scramble)
			solver := &KociembaSolver{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cube := NewCube(3)
				cube.ApplyMoves(moves)
				_, err := solver.Solve(cube)
				if err != nil {
					b.Fatalf("Solver failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkCubeOperations benchmarks core cube operations
func BenchmarkCubeOperations(b *testing.B) {
	b.Run("NewCube", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewCube(3)
		}
	})

	b.Run("IsSolved", func(b *testing.B) {
		cube := NewCube(3)
		moves, _ := ParseScramble("R U R' U'")
		cube.ApplyMoves(moves)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cube.IsSolved()
		}
	})

	b.Run("String", func(b *testing.B) {
		cube := NewCube(3)
		moves, _ := ParseScramble("R U R' U'")
		cube.ApplyMoves(moves)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cube.String()
		}
	})
}

// BenchmarkMoveOperations benchmarks move-related operations
func BenchmarkMoveOperations(b *testing.B) {
	b.Run("ParseScramble", func(b *testing.B) {
		scramble := "R U R' U' F R U R' U' F'"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ParseScramble(scramble)
		}
	})

	b.Run("ApplyMove", func(b *testing.B) {
		cube := NewCube(3)
		move := Move{Face: Right, Clockwise: true}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cube.ApplyMove(move)
		}
	})

	b.Run("ApplyMoves", func(b *testing.B) {
		moves, _ := ParseScramble("R U R' U' F R U R' U' F'")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cube := NewCube(3)
			cube.ApplyMoves(moves)
		}
	})
}

// BenchmarkSearchAlgorithms benchmarks different search strategies
func BenchmarkSearchAlgorithms(b *testing.B) {
	b.Run("BFS_2moves", func(b *testing.B) {
		moves, _ := ParseScramble("R U")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cube := NewCube(3)
			cube.ApplyMoves(moves)
			solver := &BeginnerSolver{}
			_, err := solver.Solve(cube)
			if err != nil {
				b.Fatalf("BFS failed: %v", err)
			}
		}
	})

	b.Run("BFS_4moves", func(b *testing.B) {
		moves, _ := ParseScramble("R U R' U'")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cube := NewCube(3)
			cube.ApplyMoves(moves)
			solver := &BeginnerSolver{}
			_, err := solver.Solve(cube)
			if err != nil {
				b.Fatalf("BFS failed: %v", err)
			}
		}
	})
}

// BenchmarkPatternRecognition benchmarks pattern detection
func BenchmarkPatternRecognition(b *testing.B) {
	b.Run("WhiteCross", func(b *testing.B) {
		cube := NewCube(3)
		moves, _ := ParseScramble("R U R' U'")
		cube.ApplyMoves(moves)
		pattern := &WhiteCrossPattern{}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pattern.Matches(cube)
		}
	})

	b.Run("OLLPattern", func(b *testing.B) {
		cube := NewCube(3)
		pattern := &OLLSolvedPattern{}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pattern.Matches(cube)
		}
	})

	b.Run("AnalyzeCubeState", func(b *testing.B) {
		cube := NewCube(3)
		moves, _ := ParseScramble("R U R' U'")
		cube.ApplyMoves(moves)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = AnalyzeCubeState(cube)
		}
	})
}
