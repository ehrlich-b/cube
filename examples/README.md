# 🎯 Cube Solver User Guide

Welcome to the comprehensive user guide for the Cube solver! Whether you're a complete beginner or an advanced speedcuber, this guide will help you master every feature from basic solving to advanced pattern creation.

## 🚀 Quick Start

```bash
# Build and explore your first moves
cd .. && make build
./dist/cube twist "R U R' U'" --color

# Then try solving 
./dist/cube solve "R U R' U'" --color
```

**Never used a Rubik's cube solver before?** Start with [Basic Solving](./solving.md) for a step-by-step introduction.

## 📚 Learning Path

### 🎲 Level 1: Beginner
**Start here if you're new to cubing or this solver**

- **[Basic Solving](./solving.md)** - Learn the fundamentals
  - Single moves (`R`, `U'`, `F2`)
  - Simple sequences (`R U R' U'`) 
  - Different cube sizes (2x2, 3x3, 4x4+)
  - Three solving algorithms (Beginner, CFOP, Kociemba)

### 🔥 Level 2: Intermediate  
**Ready for more advanced features**

- **[Advanced Features](./advanced.md)** - Power user tools
  - Move notation mastery (M/E/S slices, wide moves, rotations)
  - Solution verification (`cube verify`)
  - Pattern highlighting (`cube show`)  
  - Move optimization (`cube optimize`)

### 🏆 Level 3: Expert
**Master the most sophisticated features**

- **[Algorithm Database](./algorithms.md)** - Built-in algorithm lookup
  - Search famous algorithms (Sune, T-Perm, etc.)
  - Pattern recognition and previews
  - Algorithm discovery (`cube find`)

- **[Web Interface](./web.md)** - Browser-based solving
  - Terminal-style web interface
  - REST API usage
  - Integration examples

## 🎯 What Each Command Does

| Command | Purpose | Best For | Example |
|---------|---------|----------|---------|
| **`twist`** | **Apply moves and see results** | **Learning moves, exploring patterns** | **`cube twist "R U R' U'" --color`** |
| **`solve`** | Solve scrambled cubes | Learning algorithms, quick solutions | `cube solve "R U R' U'" --color` |
| **`verify`** | Check if solution works | Validating your solutions | `cube verify "R U" "U' R'" --verbose` |
| **`show`** | Display cube state | Visualizing patterns, learning | `cube show "R U R' U'" --highlight-oll` |
| **`lookup`** | Search algorithm database | Finding famous algorithms | `cube lookup sune --preview` |
| **`optimize`** | Minimize move sequences | Improving efficiency | `cube optimize "R R R"` → `R'` |
| **`find`** | Discover new algorithms | Creating custom solutions | `cube find pattern solved --max-moves 4` |
| **`serve`** | Start web interface | Browser-based usage | `cube serve --port 8080` |

## 🎨 Feature Highlights

### 🌈 Beautiful Visualization
```bash
# Compare text vs colored output
./dist/cube solve "R U R' U'"         # Clean black & white text
./dist/cube solve "R U R' U'" --color # Beautiful colored blocks 🟦🟨🟩🟧🟥⬜
```

### 🧩 Any Cube Size
```bash
./dist/cube solve "R U R' U'" --dimension 2  # 2x2 Pocket Cube
./dist/cube solve "R U R' U'" --dimension 3  # 3x3 Standard
./dist/cube solve "R U R' U'" --dimension 4  # 4x4 Revenge  
./dist/cube solve "R U R' U'" --dimension 5  # 5x5 Professor
```

### 🔄 Three Different Algorithms
```bash
./dist/cube solve "R U R' U'" --algorithm beginner  # Layer-by-layer method
./dist/cube solve "R U R' U'" --algorithm cfop      # Advanced speedcubing
./dist/cube solve "R U R' U'" --algorithm kociemba  # Optimal solutions
```

### 🎯 Advanced Notation (4x4+ cubes)
```bash
./dist/cube solve "M E S" --dimension 3 --color        # Slice moves
./dist/cube solve "Rw Fw Uw" --dimension 4 --color     # Wide moves  
./dist/cube solve "2R 3L 2F" --dimension 5 --color     # Layer moves
./dist/cube solve "x y z" --color                      # Cube rotations
```

## 🏁 Challenge Yourself

### 🎯 Famous Algorithms
```bash
# The most famous sequence in cubing
./dist/cube solve "R U R' U'" --color                           # Sexy Move

# Classic speedcubing algorithms  
./dist/cube solve "R U R' U R U2 R'" --color                    # Sune (OLL)
./dist/cube solve "R U R' F' R U R' U' R' F R2 U' R'" --color   # T-Perm (PLL)
```

### 🤯 Mind-Bending Patterns
```bash
# Create beautiful cube art
./dist/cube solve "R2 L2 U2 D2 F2 B2" --color                  # Checkerboard
./dist/cube solve "F L F U' R U F2 L2 U' L' B D' B' L2 U" --color # Cube in cube

# Ultimate mixed notation showcase (5x5)
./dist/cube solve "R M U' 2R Fw x y M' 3L Uw2 z'" --dimension 5 --color
```

### ⚡ Speed Tests
```bash
# Algorithm speed comparison
for algo in beginner cfop kociemba; do
    echo "=== $algo ==="
    time ./dist/cube solve "R U R' U' R' F R F'" --algorithm $algo
done
```

## 🎓 Learning Tips

### 🔍 Understand Before Memorizing
- Use `cube show` to visualize what each move does
- Try `cube verify` to check your understanding
- Use `--color` flag to see patterns more clearly

### 🧪 Experiment Freely  
- Start with simple moves like `R` and `U`
- Build up to sequences like `R U R' U'`
- Try the same scramble with different algorithms

### 📈 Progress Gradually
- Master 3x3 before moving to larger cubes
- Learn basic notation before advanced moves
- Use the algorithm database to discover new patterns

## 🆘 Need Help?

- **Confused by notation?** See the [Solving Guide](./solving.md#move-notation) 
- **Want to try advanced features?** Check [Advanced Features](./advanced.md)
- **Looking for specific algorithms?** Browse the [Algorithm Database](./algorithms.md)
- **Prefer web interface?** Try the [Web Guide](./web.md)

## 🎯 Ready to Dive In?

Pick your path:
- **👶 New to cubing:** [Basic Solving Guide](./solving.md)
- **🔥 Ready for more:** [Advanced Features](./advanced.md) 
- **📚 Want specific algorithms:** [Algorithm Database](./algorithms.md)
- **🌐 Prefer web interface:** [Web Interface Guide](./web.md)

**Happy cubing! 🎲✨**