# 🎲 Basic Solving Guide

Complete beginner's guide to using the Cube solver. Start here if you're new to Rubik's cubes or this solving tool.

## 🚀 Your First Moves

```bash
# Make sure you've built the project
cd .. && make build

# Start by exploring what moves do to the cube
./dist/cube twist "R U R' U'" --color
```

**What just happened?** 
- You applied the moves `R U R' U'` to a solved cube (turn Right face, then Up face, then Right counter-clockwise, then Up counter-clockwise)
- The `twist` command shows you exactly what the cube looks like after those moves
- The `--color` flag shows beautiful colored blocks instead of letters

## 🧩 Your First Solve

Now let's try solving that scrambled cube:

```bash
# Solve the scrambled state we just created
./dist/cube solve "R U R' U'" --color
```

**What happened this time?**
- The solver found a solution to return the scrambled cube back to the solved state
- You can see both the scrambled state and the solution moves

## 🔤 Understanding Move Notation

### Basic Face Moves
Each face of the cube has a letter:

| Letter | Face | Direction |
|--------|------|-----------|
| **R** | Right face | Clockwise 90° |
| **L** | Left face | Clockwise 90° |
| **U** | Up face (top) | Clockwise 90° |
| **D** | Down face (bottom) | Clockwise 90° |
| **F** | Front face | Clockwise 90° |
| **B** | Back face | Clockwise 90° |

### Modifiers
- **`'` (apostrophe)** = Counter-clockwise
- **`2`** = Double turn (180°)

### Examples to Try

```bash
# Single moves - see what each move does
./dist/cube twist "R" --color       # Right face clockwise
./dist/cube twist "R'" --color      # Right face counter-clockwise  
./dist/cube twist "R2" --color      # Right face 180° turn

# Different faces
./dist/cube twist "U" --color       # Up face
./dist/cube twist "F" --color       # Front face
./dist/cube twist "L" --color       # Left face

# Now try solving those same moves
./dist/cube solve "R" --color       # Find solution for R
./dist/cube solve "R2" --color      # Find solution for R2
```

## 🧩 Simple Sequences

### Famous Beginner Sequences

```bash
# The "Sexy Move" - most famous sequence in cubing
./dist/cube twist "R U R' U'" --color    # See what it does
./dist/cube solve "R U R' U'" --color    # Find the solution

# Right-hand algorithm (Sune)
./dist/cube twist "R U R' U R U2 R'" --color
./dist/cube solve "R U R' U R U2 R'" --color

# Left-hand algorithm  
./dist/cube twist "L' U' L U' L' U2 L" --color
./dist/cube solve "L' U' L U' L' U2 L" --color

# Simple triggers
./dist/cube twist "R U R'" --color       # See the pattern
./dist/cube solve "R U R'" --color       # Find solution
```

### Building Longer Sequences

```bash
# Combine simple moves - first explore, then solve
./dist/cube twist "R U" --color              # Right then Up
./dist/cube solve "R U" --color              # Find solution
./dist/cube solve "R U R'" --color           # Right, Up, Right back
./dist/cube solve "R U R' U'" --color        # The sexy move
./dist/cube solve "R U R' U' R U R' U'" --color  # Double sexy move
```

## 🎯 Different Cube Sizes

The solver supports cubes from 2x2 up to 10x10 and beyond!

### 2x2x2 Pocket Cube
```bash
# Smaller cube, simpler moves
./dist/cube solve "R U R' U'" --dimension 2 --color

# Classic 2x2 sequences
./dist/cube solve "R U R' F R F'" --dimension 2 --color
./dist/cube solve "R U2 R' U' R U' R'" --dimension 2 --color
```

### 3x3x3 Standard Cube
```bash
# This is the default dimension
./dist/cube solve "R U R' U'" --color
./dist/cube solve "R U R' U'" --dimension 3 --color  # Same thing

# Famous 3x3 algorithms
./dist/cube solve "R U R' U R U2 R'" --color                    # Sune
./dist/cube solve "R U R' F' R U R' U' R' F R2 U' R'" --color   # T-Perm
```

### 4x4x4 Revenge Cube  
```bash
# Larger cube, same basic moves
./dist/cube solve "R U R' U'" --dimension 4 --color

# 4x4 handles more complex scrambles
./dist/cube solve "R U R' U' R' F R F'" --dimension 4 --color
```

### 5x5x5 Professor Cube
```bash
# Even larger!
./dist/cube solve "R U R' U'" --dimension 5 --color

# More complex patterns possible
./dist/cube solve "R U2 R' U' R U' R'" --dimension 5 --color
```

## 🤖 Three Solving Algorithms

Our solver has three different algorithms that solve cubes in different ways:

### 1. Beginner Algorithm
**Best for**: Learning, understanding how cubes work
**Method**: Layer-by-layer approach

```bash
./dist/cube solve "R U R' U'" --algorithm beginner --color
```

### 2. CFOP Algorithm  
**Best for**: Speedcubing, advanced users
**Method**: Cross → F2L → OLL → PLL (the method used by speedcubers)

```bash
./dist/cube solve "R U R' U'" --algorithm cfop --color
```

### 3. Kociemba Algorithm
**Best for**: Optimal solutions, mathematical interest  
**Method**: Two-phase algorithm for efficient solutions

```bash
./dist/cube solve "R U R' U'" --algorithm kociemba --color
```

### Compare All Three
```bash
# See how each algorithm solves the same scramble differently
echo "=== BEGINNER ==="
./dist/cube solve "R U R' U' R' F R F'" --algorithm beginner --color

echo "=== CFOP ==="  
./dist/cube solve "R U R' U' R' F R F'" --algorithm cfop --color

echo "=== KOCIEMBA ==="
./dist/cube solve "R U R' U' R' F R F'" --algorithm kociemba --color
```

## 🎨 Visual Options

### Text vs Color Output

```bash
# Clean black and white (default)
./dist/cube solve "R U R' U'"

# Beautiful colored blocks (recommended!)
./dist/cube solve "R U R' U'" --color
```

The color output uses:
- 🟦 **Blue** for Blue face
- 🟨 **Yellow** for Yellow face  
- 🟩 **Green** for Green face
- 🟧 **Orange** for Orange face
- 🟥 **Red** for Red face
- ⬜ **White** for White face

## 🧪 Practice Exercises

### Exercise 1: Single Moves
Try each face move and see what happens:

```bash
./dist/cube solve "R" --color     # Right
./dist/cube solve "L" --color     # Left  
./dist/cube solve "U" --color     # Up
./dist/cube solve "D" --color     # Down
./dist/cube solve "F" --color     # Front
./dist/cube solve "B" --color     # Back
```

### Exercise 2: Modifiers
Learn the modifiers by practicing:

```bash
./dist/cube solve "R" --color     # Clockwise
./dist/cube solve "R'" --color    # Counter-clockwise
./dist/cube solve "R2" --color    # Double turn
```

### Exercise 3: Building Sequences
Start simple and build up:

```bash
./dist/cube solve "R U" --color           # Two moves
./dist/cube solve "R U R'" --color        # Three moves  
./dist/cube solve "R U R' U'" --color     # Four moves (sexy move)
```

### Exercise 4: Different Cubes
Try the same sequence on different cube sizes:

```bash
./dist/cube solve "R U R' U'" --dimension 2 --color   # 2x2
./dist/cube solve "R U R' U'" --dimension 3 --color   # 3x3
./dist/cube solve "R U R' U'" --dimension 4 --color   # 4x4
```

## 🎯 Cool Patterns to Try

### Beginner-Friendly Patterns

```bash
# Checkerboard pattern
./dist/cube solve "R2 L2 U2 D2 F2 B2" --color

# Four spots pattern  
./dist/cube solve "F R U' R' F' R U R'" --color

# Cross pattern
./dist/cube solve "F B' R L F B' R L" --color
```

### Classic Algorithms

```bash
# Sune - most common OLL case
./dist/cube solve "R U R' U R U2 R'" --color

# Antisune
./dist/cube solve "R U2 R' U' R U' R'" --color

# T-Perm - swaps two adjacent corners
./dist/cube solve "R U R' F' R U R' U' R' F R2 U' R'" --color
```

## 🔍 Understanding Solutions

The solver shows you:
- **Scramble**: The moves that mixed up the cube
- **Solution**: The moves needed to solve it
- **Steps**: How many moves in the solution
- **Duration**: How long the algorithm took

```bash
# Example output:
# Solving 3x3x3 cube...
# Using algorithm: beginner
# Solution: U R U' R'  
# Solved in 4 steps (took 1.2ms)
```

## 🚨 Common Beginner Mistakes

### 1. Mixing up `'` and `2`
- `R'` = Right counter-clockwise (90°)
- `R2` = Right double turn (180°)

### 2. Forgetting spaces in sequences
```bash
# ❌ Wrong - no spaces
./dist/cube solve "RUR'U'" 

# ✅ Correct - spaces between moves
./dist/cube solve "R U R' U'" 
```

### 3. Case sensitivity
```bash
# ❌ Wrong - lowercase
./dist/cube solve "r u r' u'"

# ✅ Correct - uppercase
./dist/cube solve "R U R' U'"
```

## ⚡ Quick Reference

### Essential Commands
```bash
# Basic solve with colors
./dist/cube solve "SCRAMBLE" --color

# Different algorithms
./dist/cube solve "SCRAMBLE" --algorithm beginner
./dist/cube solve "SCRAMBLE" --algorithm cfop  
./dist/cube solve "SCRAMBLE" --algorithm kociemba

# Different cube sizes
./dist/cube solve "SCRAMBLE" --dimension 2
./dist/cube solve "SCRAMBLE" --dimension 3
./dist/cube solve "SCRAMBLE" --dimension 4
```

### Face Letters
- **R** = Right, **L** = Left
- **U** = Up, **D** = Down  
- **F** = Front, **B** = Back

### Modifiers
- **`'`** = Counter-clockwise
- **`2`** = Double turn (180°)

## 🎯 Next Steps

Once you're comfortable with basic solving:

1. **Try advanced features**: Check out [Advanced Features](./advanced.md)
2. **Learn specific algorithms**: Browse the [Algorithm Database](./algorithms.md) 
3. **Use the web interface**: Try the [Web Interface Guide](./web.md)

**Ready for more challenge?** Move on to [Advanced Features](./advanced.md) to learn about wide moves, slice moves, and power user tools!