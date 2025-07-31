# CFEN: Cube Forsyth-Edwards Notation

**A standardized, compact text format for representing Rubik's cube states with wildcard support and NxN scalability.**

## Overview

CFEN (Cube Forsyth-Edwards Notation) is inspired by chess FEN notation and provides a single-line, human-readable format for describing any Rubik's cube state. Unlike move sequences, CFEN captures the **exact visual state** of a cube, including orientation and partial/unknown positions.

### Why CFEN?

Traditional cube tools only exchange move sequences (`R U R' U'`), but this approach has critical limitations:

- **No orientation info**: Same moves, different results depending on starting orientation
- **No partial states**: Can't express "white cross solved, rest unknown"  
- **No cube size flexibility**: Move sequences don't scale cleanly to 4x4, 5x5, etc.
- **No verification format**: Hard to specify "I only care about these specific positions"

CFEN solves all these problems with a chess FEN-inspired approach.

## Format Structure

```
<orientation>|<faces>
```

### Orientation Field: `<Up><Front>`
Two letters specify which colors currently face Up and Front:
- `WG` = White up, Green front (standard orientation)
- `YB` = Yellow up, Blue front (cube rotated)

### Faces Field: `<U>/<R>/<F>/<D>/<L>/<B>`
Six face descriptions separated by `/`, in fixed order:
- **U**p face (currently facing up)
- **R**ight face  
- **F**ront face (currently facing forward)
- **D**own face
- **L**eft face
- **B**ack face

## Alphabet & Tokens

| Token | Meaning |
|-------|---------|
| `W` `Y` `R` `O` `G` `B` | Fixed sticker colors (White, Yellow, Red, Orange, Green, Blue) |
| `?` | Wildcard - color and position don't matter |
| `0-9` | Run-length digits (placed immediately after a token) |
| `/` | Face separator |
| `|` | Orientation/faces separator |

## Run-Length Encoding

To avoid repetition, CFEN uses run-length encoding:

| Notation | Expands To | Meaning |
|----------|------------|---------|
| `W` | `W` | One white sticker |
| `W9` | `WWWWWWWWW` | Nine white stickers |
| `?5` | `?????` | Five wildcard positions |
| `R16` | `RRRRRRRRRRRRRRRR` | Sixteen red stickers |

**Rules:**
- Digits apply to the **immediately preceding** token
- Single stickers can omit the `1` (`W` = `W1`)
- Leading zeros forbidden (`?10` not `?010`)

## Cube Size Detection

CFEN automatically determines cube dimensions:
1. Count stickers per face after expanding run-length
2. Each face must have the same count `S`
3. `S` must be a perfect square
4. Cube size `N = √S`

Examples:
- 9 stickers per face → 3×3 cube
- 16 stickers per face → 4×4 cube  
- 25 stickers per face → 5×5 cube

## Practical Examples

### Solved 3×3 Cube
```
WG|WWWWWWWWW/RRRRRRRRR/GGGGGGGGG/YYYYYYYYY/OOOOOOOOO/BBBBBBBBB
```
Standard orientation, all faces solid colors.

### 3×3 White Cross Only
```
WG|?W?WWW?W?/?9/?9/?9/?9/?9
```
- Up face: wildcards except cross pattern (center + 4 edges)
- All other faces: completely unknown (`?9` = 9 wildcards)

### Solved 4×4 Cube  
```
WG|W16/R16/G16/Y16/O16/B16
```
Each face has 16 stickers, all same color.

### 4×4 White Cross Only
```
WG|?W??W??W??W?/?16/?16/?16/?16/?16
```
- Up face: specific cross pattern for 4×4
- Other faces: all wildcards

### OLL Practice (5×5, Yellow Top Layer)
```
WG|Y25/?25/?25/?25/?25/?25
```
- Up face: all 25 positions are yellow (OLL solved)
- Other faces: don't care (all wildcards)

### Massive Cube (17×17, Only UF Edge Fixed)
```
WG|?289/?289/?289/?289/?289/?289  
```
For algorithm development on huge cubes where only specific pieces matter.

## Implementation Benefits

### For Layer-by-Layer Solving
```bash
# Step 1: White cross verification
cube verify-cfen "scramble" "solution" --target "WG|?W?WWW?W?/?9/?9/?9/?9/?9"

# Step 2: White face verification  
cube verify-cfen "scramble" "solution" --target "WG|W9/?9/?9/?9/?9/?9"

# Step 3: Middle layer verification
cube verify-cfen "scramble" "solution" --target "WG|W9/?W?W?W?W?/?9/?9/?9/?9"
```

### For Algorithm Testing
```bash
# Test Sune algorithm on any OLL case where top isn't all yellow
cube test-algorithm "R U R' U R U2 R'" --input "WG|Y?Y?Y?Y?Y/?9/?9/?9/?9/?9"
```

### For Multi-Dimensional Support
```bash
# Same verification logic works for any cube size
cube verify-cfen "4x4-scramble" "solution" --target "WG|W16/?16/?16/?16/?16/?16"
```

## Validation Rules

1. **Orientation field**: Exactly 2 color letters
2. **Face count**: Exactly 6 faces separated by `/`
3. **Face consistency**: All faces must expand to same sticker count
4. **Perfect square**: Sticker count must be N² for some integer N
5. **Total stickers**: Must equal 6×N²
6. **Valid tokens**: Only `WYROGB?` and digits `0-9`

## Reference Implementation Notes

A CFEN parser should:
1. Split on `|` to separate orientation and faces
2. Split faces on `/` (should get exactly 6)
3. Expand each face using run-length rules
4. Validate all faces have same sticker count
5. Verify sticker count is perfect square
6. Extract cube dimension N = √(stickers per face)

## Why This Solves Our Problems

✅ **FEN-equivalent**: Single-line, diff-friendly, unambiguous  
✅ **Orientation-aware**: Explicit Up/Front specification  
✅ **Wildcard support**: `?` for "don't care" positions  
✅ **NxN scalable**: Works for 2×2 through 17×17+ cubes  
✅ **Partial states**: Perfect for layer-by-layer algorithm verification  
✅ **Human-readable**: Can type `WG|W9/?9/?9/?9/?9/?9` by hand  
✅ **Compact**: Run-length encoding prevents massive strings  

This format gives us everything we need to build a proper cube solver with grey square verification system.