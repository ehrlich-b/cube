package cube

// Ring generators define which stickers move for each type of move

// ringR generates ring coordinates for R move at layer k
func ringR(N, k int) []Coord {
	var ring []Coord
	// Up face: column N-1-k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Up, r, N - 1 - k})
	}
	// Back face: column k, rows N-1 to 0 (opposite direction due to 3D orientation)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Back, r, k})
	}
	// Down face: column N-1-k, rows 0 to N-1 (same as Up, they're parallel)
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Down, r, N - 1 - k})
	}
	// Front face: column N-1-k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Front, r, N - 1 - k})
	}
	return ring
}

// ringL generates ring coordinates for L move at layer k
func ringL(N, k int) []Coord {
	var ring []Coord
	// Up face: column k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Up, r, k})
	}
	// Front face: column k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Front, r, k})
	}
	// Down face: column k, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Down, r, k})
	}
	// Back face: column N-1-k, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Back, r, N - 1 - k})
	}
	return ring
}

// ringU generates ring coordinates for U move at layer k
func ringU(N, k int) []Coord {
	var ring []Coord
	// Back face: row k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Back, k, c})
	}
	// Right face: row k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Right, k, c})
	}
	// Front face: row k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Front, k, c})
	}
	// Left face: row k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Left, k, c})
	}
	return ring
}

// ringD generates ring coordinates for D move at layer k
func ringD(N, k int) []Coord {
	var ring []Coord
	// Front face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Front, N - 1 - k, c})
	}
	// Right face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Right, N - 1 - k, c})
	}
	// Back face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Back, N - 1 - k, c})
	}
	// Left face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Left, N - 1 - k, c})
	}
	return ring
}

// ringF generates ring coordinates for F move at layer k
func ringF(N, k int) []Coord {
	var ring []Coord
	// Up face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Up, N - 1 - k, c})
	}
	// Right face: column k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Right, r, k})
	}
	// Down face: row k, columns N-1 to 0 (reversed)
	for c := N - 1; c >= 0; c-- {
		ring = append(ring, Coord{Down, k, c})
	}
	// Left face: column N-1-k, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Left, r, N - 1 - k})
	}
	return ring
}

// ringB generates ring coordinates for B move at layer k
func ringB(N, k int) []Coord {
	var ring []Coord
	// Up face: row k, columns N-1 to 0 (reversed)
	for c := N - 1; c >= 0; c-- {
		ring = append(ring, Coord{Up, k, c})
	}
	// Left face: column k, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Left, r, k})
	}
	// Down face: row N-1-k, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Down, N - 1 - k, c})
	}
	// Right face: column N-1-k, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Right, r, N - 1 - k})
	}
	return ring
}

// Slice move ring generators

// ringM generates ring coordinates for M slice move (between L and R)
func ringM(N, k int) []Coord {
	if N%2 == 0 {
		return nil // M slice undefined for even cubes for now
	}
	centerCol := N / 2
	var ring []Coord
	// Up face: column centerCol, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Up, r, centerCol})
	}
	// Front face: column centerCol, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Front, r, centerCol})
	}
	// Down face: column centerCol, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Down, r, centerCol})
	}
	// Back face: column centerCol, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Back, r, centerCol})
	}
	return ring
}

// ringE generates ring coordinates for E slice move (between U and D)
func ringE(N, k int) []Coord {
	if N%2 == 0 {
		return nil // E slice undefined for even cubes for now
	}
	centerRow := N / 2
	var ring []Coord
	// Front face: row centerRow, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Front, centerRow, c})
	}
	// Left face: row centerRow, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Left, centerRow, c})
	}
	// Back face: row centerRow, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Back, centerRow, c})
	}
	// Right face: row centerRow, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Right, centerRow, c})
	}
	return ring
}

// ringS generates ring coordinates for S slice move (between F and B)
func ringS(N, k int) []Coord {
	if N%2 == 0 {
		return nil // S slice undefined for even cubes for now
	}
	centerSlice := N / 2
	var ring []Coord
	// Up face: row centerSlice, columns 0 to N-1
	for c := 0; c < N; c++ {
		ring = append(ring, Coord{Up, centerSlice, c})
	}
	// Right face: column centerSlice, rows 0 to N-1
	for r := 0; r < N; r++ {
		ring = append(ring, Coord{Right, r, centerSlice})
	}
	// Down face: row centerSlice, columns N-1 to 0 (reversed)
	for c := N - 1; c >= 0; c-- {
		ring = append(ring, Coord{Down, centerSlice, c})
	}
	// Left face: column centerSlice, rows N-1 to 0 (reversed)
	for r := N - 1; r >= 0; r-- {
		ring = append(ring, Coord{Left, r, centerSlice})
	}
	return ring
}
