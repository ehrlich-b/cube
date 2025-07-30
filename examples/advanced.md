# 🔥 Advanced Features Guide

Master the sophisticated features that make this cube solver truly powerful. These tools are designed for serious cubers and power users.

## 🎯 Prerequisites

Make sure you've mastered the basics first:
- Comfortable with basic moves (`R`, `U'`, `F2`)
- Understand the three solving algorithms
- Know how to use different cube dimensions

**New to the basics?** Start with the [Basic Solving Guide](./solving.md) first.

## 🚀 Advanced Move Notation

Beyond basic face moves lies a whole world of advanced notation for complex manipulations.

### 🔄 Middle Slice Moves (3x3, 5x5, 7x7...)

Middle slices only exist on odd-sized cubes (3x3, 5x5, 7x7, etc.).

```bash
# M-slice: Middle slice between Left and Right faces
./dist/cube solve "M" --dimension 3 --color     # M-slice clockwise
./dist/cube solve "M'" --dimension 3 --color    # M-slice counter-clockwise
./dist/cube solve "M2" --dimension 3 --color    # M-slice double turn

# E-slice: Equatorial slice between Up and Down faces
./dist/cube solve "E" --dimension 3 --color     # E-slice clockwise
./dist/cube solve "E'" --dimension 3 --color    # E-slice counter-clockwise
./dist/cube solve "E2" --dimension 3 --color    # E-slice double turn

# S-slice: Standing slice between Front and Back faces
./dist/cube solve "S" --dimension 3 --color     # S-slice clockwise
./dist/cube solve "S'" --dimension 3 --color    # S-slice counter-clockwise
./dist/cube solve "S2" --dimension 3 --color    # S-slice double turn
```

### 🔄 All Slice Moves Together

```bash
# Try all three slice moves
./dist/cube solve "M E S" --dimension 3 --color

# Complex slice combinations
./dist/cube solve "M' E2 S'" --dimension 3 --color
./dist/cube solve "M2 E M2 E2 M2 E M2" --dimension 3 --color

# Slice moves on 5x5 cube
./dist/cube solve "M E' S M' E S'" --dimension 5 --color
```

### 🌊 Wide Moves (4x4+ cubes)

Wide moves affect multiple layers at once - perfect for larger cubes.

```bash
# Basic wide moves (affects outer 2 layers)
./dist/cube solve "Rw" --dimension 4 --color    # Right wide
./dist/cube solve "Lw" --dimension 4 --color    # Left wide  
./dist/cube solve "Uw" --dimension 4 --color    # Up wide
./dist/cube solve "Dw" --dimension 4 --color    # Down wide
./dist/cube solve "Fw" --dimension 4 --color    # Front wide
./dist/cube solve "Bw" --dimension 4 --color    # Back wide

# Wide moves with modifiers
./dist/cube solve "Rw'" --dimension 4 --color   # Right wide counter-clockwise
./dist/cube solve "Uw2" --dimension 4 --color   # Up wide double turn
```

### 🌊 Wide Move Combinations

```bash
# All wide moves together
./dist/cube solve "Rw Lw Uw Dw Fw Bw" --dimension 4 --color

# Complex wide patterns
./dist/cube solve "Rw U2 Rw' U2 Rw U2 Rw'" --dimension 4 --color

# Wide moves on 5x5
./dist/cube solve "Rw2 Uw Rw2 Uw2 Rw2 Uw Rw2" --dimension 5 --color
```

### 🎯 Layer-Specific Moves (4x4+)

Target specific inner layers with numbered notation.

```bash
# 4x4 cube - inner layer moves
./dist/cube solve "2R" --dimension 4 --color    # Second layer from right
./dist/cube solve "2L" --dimension 4 --color    # Second layer from left
./dist/cube solve "2U" --dimension 4 --color    # Second layer from up
./dist/cube solve "2F" --dimension 4 --color    # Second layer from front

# 5x5 cube - multiple inner layers
./dist/cube solve "2R 3L" --dimension 5 --color             # 2nd from right, 3rd from left
./dist/cube solve "2R 2L 2U 2D 2F 2B" --dimension 5 --color # All second layers

# 6x6 cube - even more layers
./dist/cube solve "2R 3R 2L 3L" --dimension 6 --color       # Multiple inner layers
./dist/cube solve "2R 3R' 2L' 3L" --dimension 6 --color     # With different directions
```

### 🌍 Cube Rotations

Rotate the entire cube in space - doesn't change the solution, just your perspective.

```bash
# Individual rotations
./dist/cube solve "x" --dimension 3 --color     # Rotate along R-L axis (like doing R)
./dist/cube solve "y" --dimension 3 --color     # Rotate along U-D axis (like doing U)  
./dist/cube solve "z" --dimension 3 --color     # Rotate along F-B axis (like doing F)

# Rotations with modifiers
./dist/cube solve "x'" --dimension 3 --color    # Counter-clockwise rotation
./dist/cube solve "y2" --dimension 3 --color    # Double rotation (180°)

# Multiple rotations
./dist/cube solve "x y z" --dimension 3 --color
./dist/cube solve "x2 y' z2" --dimension 3 --color
```

### 🤯 Ultimate Mixed Notation

Combine everything for maximum complexity:

```bash
# 5x5 cube with every notation type
./dist/cube solve "R M U' 2R Fw x y M' 3L Uw2 z'" --dimension 5 --color

# 4x4 symmetric madness
./dist/cube solve "Rw Lw' Uw Dw' Fw Bw' x y' z" --dimension 4 --color

# 6x6 layer chaos
./dist/cube solve "2R 3L 2U 3D 2F 3B Rw Lw' x2 y z'" --dimension 6 --color

# Everything at once (5x5)
./dist/cube solve "R M E S U' R' E S Rw Uw 2R 2L x y' z 3R 3L'" --dimension 5 --color
```

## ✅ Solution Verification

Verify that your solutions actually work with the `cube verify` command.

### Basic Verification

```bash
# Verify a correct solution
./dist/cube verify "R U R' U'" "U R U' R'"

# Verify an incorrect solution (will fail)
./dist/cube verify "R U R' U'" "F U F' U'"
```

### Verbose Verification

See exactly what happens step by step:

```bash
# Show detailed verification process
./dist/cube verify "R U" "U' R'" --verbose

# With colors for better visualization
./dist/cube verify "R U" "U' R'" --verbose --color
```

### Verification on Different Cube Sizes

```bash
# Verify 2x2 solution
./dist/cube verify "R U R' U'" "U R U' R'" --dimension 2

# Verify 4x4 solution with wide moves
./dist/cube verify "Rw" "Rw'" --dimension 4 --verbose --color

# Verify 5x5 solution with layer moves
./dist/cube verify "2R 3L" "3L' 2R'" --dimension 5 --verbose
```

### Testing Your Understanding

```bash
# These should work (correct inverse pairs)
./dist/cube verify "R" "R'" --verbose
./dist/cube verify "R2" "R2" --verbose  
./dist/cube verify "R U R' U'" "U R U' R'" --verbose

# These should fail (incorrect solutions)
./dist/cube verify "R" "U" --verbose
./dist/cube verify "R U R' U'" "R U R' U'" --verbose
```

## 🎨 Pattern Highlighting

Use `cube show` to display cube states with pattern highlighting.

### Basic Display

```bash
# Show solved cube
./dist/cube show

# Show scrambled cube
./dist/cube show "R U R' U'" --color

# Show complex scramble
./dist/cube show "R U R' F' R U R' U' R' F R2 U' R'" --color
```

### Pattern Highlighting

Highlight specific patterns to understand cube states:

```bash
# Highlight cross pattern (yellow edges on top)
./dist/cube show "R U R' U'" --highlight-cross --color

# Highlight OLL pattern (orientation of last layer)
./dist/cube show "R U R' U'" --highlight-oll --color

# Highlight PLL pattern (permutation of last layer)  
./dist/cube show "R U R' U'" --highlight-pll --color

# Highlight F2L pattern (first two layers)
./dist/cube show "R U R' U'" --highlight-f2l --color
```

### Progressive Pattern Analysis

Watch how patterns develop:

```bash
# Start with cross highlighting
./dist/cube show "F R U' R' F'" --highlight-cross --color

# Then show OLL development
./dist/cube show "R U R' U R U2 R'" --highlight-oll --color

# Finally PLL patterns
./dist/cube show "R U R' F' R U R' U' R' F R2 U' R'" --highlight-pll --color
```

## ⚡ Move Optimization

Use `cube optimize` to minimize move sequences and improve efficiency.

### Basic Optimization

```bash
# Combine consecutive moves
./dist/cube optimize "R R"              # → R2 (1 move instead of 2)
./dist/cube optimize "R R R"            # → R' (1 move instead of 3)  
./dist/cube optimize "R R R R"          # → (empty - all moves cancel)

# Remove canceling moves
./dist/cube optimize "R R'"             # → (empty - moves cancel)
./dist/cube optimize "R' R"             # → (empty - moves cancel)
./dist/cube optimize "R2 R2"            # → (empty - moves cancel)
```

### Advanced Optimization

```bash
# Complex sequences
./dist/cube optimize "R R U U' F F F"   # → R2 F' (3 moves instead of 7)
./dist/cube optimize "R U R R U' F F'"  # → R U R2 U' (4 moves instead of 7)

# Mixed notation optimization
./dist/cube optimize "Rw Rw"            # → Rw2
./dist/cube optimize "2R 2R 2R"         # → 2R'
```

### Real-World Optimization

```bash
# Optimize famous algorithms
./dist/cube optimize "R U R' U R U R' U R U R' U'"  # Repetitive sequence
./dist/cube optimize "R U R' U' R U R' U' R U R' U'" # Another pattern

# Optimize your own sequences
./dist/cube optimize "R R U' U' R R' R' U U"        # Your custom algorithm
```

## 🔍 Algorithm Discovery

Use `cube find` to discover new algorithms and solutions.

### Pattern-Based Discovery

```bash
# Find sequences that solve the cube
./dist/cube find pattern solved --max-moves 4

# Find sequences that create a cross pattern
./dist/cube find pattern cross --max-moves 6 --from "R U"

# Find solutions starting from a specific state
./dist/cube find pattern solved --max-moves 5 --from "R U R' U'"
```

### Sequence-Based Discovery

```bash
# Find solutions to specific scrambles
./dist/cube find sequence "R U" --max-moves 4

# Find solutions to complex scrambles
./dist/cube find sequence "R U R' U'" --max-moves 6

# Discover multiple solutions
./dist/cube find sequence "R" --max-moves 3    # Should find R', R R R, etc.
```

### Advanced Discovery Options

```bash
# Show step-by-step discovery process
./dist/cube find pattern solved --max-moves 4 --steps

# Find longer solutions
./dist/cube find pattern cross --max-moves 8 --from "F"

# Discover solutions from complex starting positions
./dist/cube find pattern solved --max-moves 6 --from "R U R' F' R U R' U' R' F R2 U' R'"
```

## 🎯 Power User Workflows

### Complete Solution Analysis

```bash
# 1. Start with a scramble
SCRAMBLE="R U R' U' R' F R F'"

# 2. Solve it
./dist/cube solve "$SCRAMBLE" --algorithm beginner --color

# 3. Get the solution and verify it works
SOLUTION=$(./dist/cube solve "$SCRAMBLE" --algorithm beginner | grep "Solution:" | sed 's/Solution: //')
./dist/cube verify "$SCRAMBLE" "$SOLUTION" --verbose --color

# 4. Optimize the solution
./dist/cube optimize "$SOLUTION"

# 5. Visualize the final state
./dist/cube show "$SCRAMBLE" --highlight-oll --color
```

### Algorithm Development Workflow

```bash
# 1. Find a short solution to a specific scramble
./dist/cube find sequence "R U R'" --max-moves 4

# 2. Verify the discovered solution works
./dist/cube verify "R U R'" "R' U' R'" --verbose

# 3. Optimize the solution
./dist/cube optimize "R' U' R'"

# 4. Test on different cube sizes
./dist/cube verify "R U R'" "R' U' R'" --dimension 4 --verbose
```

### Pattern Research Workflow

```bash
# 1. Create an interesting pattern
./dist/cube show "R2 L2 U2 D2 F2 B2" --color

# 2. Analyze what patterns are highlighted
./dist/cube show "R2 L2 U2 D2 F2 B2" --highlight-cross --color
./dist/cube show "R2 L2 U2 D2 F2 B2" --highlight-oll --color

# 3. Find the shortest way to create this pattern
./dist/cube find sequence "R2 L2 U2 D2 F2 B2" --max-moves 8

# 4. Optimize the result
PATTERN_SOLUTION=$(./dist/cube find sequence "R2 L2 U2 D2 F2 B2" --max-moves 8 | head -1 | cut -d' ' -f2-)
./dist/cube optimize "$PATTERN_SOLUTION"
```

## 🎮 Advanced Challenges

### Challenge 1: Master All Notation Types
Try to create a sequence using every notation type:

```bash
# Your goal: Use R, M, Rw, 2R, and x in one sequence
./dist/cube solve "R M Rw 2R x U' M' Rw' 2R' x'" --dimension 5 --color
```

### Challenge 2: Optimize Everything  
Take a long, inefficient sequence and make it as short as possible:

```bash
# Start with this mess
./dist/cube optimize "R R U U U' U' U R' R' R F F' F' F R R'"
```

### Challenge 3: Discover Your Own Algorithm
Find a 4-move solution to `R U R'`:

```bash
# Should find R' U' R' and other solutions
./dist/cube find sequence "R U R'" --max-moves 4
```

### Challenge 4: Cross-Dimension Verification
Test if the same solution works on different cube sizes:

```bash
SOLUTION="R' U' R'"
./dist/cube verify "R U R'" "$SOLUTION" --dimension 3 --verbose
./dist/cube verify "R U R'" "$SOLUTION" --dimension 4 --verbose  
./dist/cube verify "R U R'" "$SOLUTION" --dimension 5 --verbose
```

## 📚 Quick Reference

### Advanced Notation
- **M/E/S**: Middle slices (odd cubes only)
- **Rw/Lw/Uw/Dw/Fw/Bw**: Wide moves (4x4+)
- **2R/3L/4U**: Layer-specific moves (4x4+)
- **x/y/z**: Cube rotations

### Power Tools
- **`verify`**: Check if solutions work
- **`show`**: Display cube states with pattern highlighting
- **`optimize`**: Minimize move sequences
- **`find`**: Discover new algorithms

### Useful Flags
- **`--verbose`**: Show detailed step-by-step process
- **`--color`**: Beautiful colored visualization
- **`--max-moves N`**: Limit search depth
- **`--steps`**: Show intermediate steps

## 🎯 Next Steps

Ready to dive even deeper?

1. **Master specific algorithms**: Browse the [Algorithm Database](./algorithms.md)
2. **Try the web interface**: Check out the [Web Interface Guide](./web.md)
3. **Contribute your discoveries**: Share your optimized sequences and new algorithms!

**Feeling confident with advanced features?** You're now ready to explore the [Algorithm Database](./algorithms.md) and discover the vast world of speedcubing algorithms!