package cube

// SliceType represents middle slice moves
type SliceType int

const (
	NoSlice SliceType = iota
	M_Slice           // Between L and R faces
	E_Slice           // Between U and D faces
	S_Slice           // Between F and B faces
)

// RotationType represents cube rotations
type RotationType int

const (
	NoRotation RotationType = iota
	X_Rotation              // Rotation around R face axis
	Y_Rotation              // Rotation around U face axis
	Z_Rotation              // Rotation around F face axis
)

// Move represents a single move
type Move struct {
	Face      Face         // Which face to turn (R, L, U, D, F, B)
	Clockwise bool         // True for clockwise, false for counter-clockwise
	Double    bool         // True for 180-degree turns
	Wide      bool         // True for wide turns (Rw, Uw, etc.)
	WideDepth int          // How many layers for numbered wide turns (2Rw = depth 2)
	Layer     int          // For numbered layer turns (2R, 3L, etc.)
	Slice     SliceType    // For slice turns (M, E, S)
	Rotation  RotationType // For cube rotations (x, y, z)
}

// MoveType represents different types of moves for permutation generation
type MoveType int

const (
	MoveR MoveType = iota
	MoveL
	MoveU
	MoveD
	MoveF
	MoveB
	MoveM // M slice
	MoveE // E slice
	MoveS // S slice
	MoveX // x rotation
	MoveY // y rotation
	MoveZ // z rotation
)

// Coord represents a sticker coordinate
type Coord struct {
	Face Face
	Row  int
	Col  int
}

// stickerIndex converts (face, row, col) to flat index
func stickerIndex(face Face, row, col, N int) int {
	return int(face)*N*N + row*N + col
}

// indexToCoord converts flat index back to (face, row, col)
func indexToCoord(idx, N int) (Face, int, int) {
	face := Face(idx / (N * N))
	remainder := idx % (N * N)
	row := remainder / N
	col := remainder % N
	return face, row, col
}

// Permutation represents a mapping of sticker indices
type Permutation []int
