# 📚 Algorithm Database Guide

Discover and master the built-in collection of famous speedcubing algorithms. Search by name, pattern, or category to find exactly what you need.

## 🎯 Overview

The cube solver includes a database of **15 essential algorithms** across all major speedcubing categories:
- **5 OLL algorithms** (Orientation of Last Layer)
- **6 PLL algorithms** (Permutation of Last Layer)  
- **2 F2L algorithms** (First Two Layers)
- **2 Common triggers** (Fundamental patterns)

## 🔍 Basic Algorithm Lookup

### Search by Name

```bash
# Find the famous Sune algorithm
./dist/cube lookup sune

# Search for T-Perm
./dist/cube lookup "T-Perm"

# Find the Sexy Move
./dist/cube lookup "sexy move"
```

### Search by Pattern

```bash
# Find algorithms containing specific moves
./dist/cube lookup --pattern "R U R' U'"

# Search for algorithms with M moves
./dist/cube lookup --pattern "M"

# Find F2L patterns
./dist/cube lookup --pattern "U' R'"
```

### Search by Category

```bash
# Show all OLL algorithms
./dist/cube lookup --category OLL

# Show all PLL algorithms  
./dist/cube lookup --category PLL

# Show F2L algorithms
./dist/cube lookup --category F2L

# Show trigger moves
./dist/cube lookup --category Trigger
```

### Show All Algorithms

```bash
# Display the complete database
./dist/cube lookup --all
```

## 🎯 Algorithm Categories

### 🌟 OLL (Orientation of Last Layer)

These algorithms orient the pieces on the top layer to create a solid color on top.

```bash
# Sune - Most common OLL case
./dist/cube lookup sune --preview
# Algorithm: R U R' U R U2 R'

# Anti-Sune - Mirror of Sune
./dist/cube lookup "anti-sune" --preview  
# Algorithm: R U2 R' U' R U' R'

# Cross OLL - Creates cross pattern
./dist/cube lookup "cross oll" --preview
# Algorithm: F R U R' U' F'

# Dot OLL - No edges oriented
./dist/cube lookup "dot oll" --preview
# Algorithm: F R U R' U' F' f R U R' U' f'

# L-Shape OLL - L-shaped edge pattern
./dist/cube lookup "l-shape" --preview
# Algorithm: F U R U' R' F'
```

### 🔄 PLL (Permutation of Last Layer)

These algorithms move pieces around while keeping the top layer oriented.

```bash
# T-Perm - Most common PLL case
./dist/cube lookup "t-perm" --preview
# Algorithm: R U R' U' R' F R2 U' R' U' R U R' F'

# Y-Perm - Diagonal corner swap
./dist/cube lookup "y-perm" --preview
# Algorithm: F R U' R' U' R U R' F' R U R' U' R' F R F'

# U-Perm (a) - Edge 3-cycle counterclockwise
./dist/cube lookup "u-perm (a)" --preview
# Algorithm: R U' R U R U R U' R' U' R2

# U-Perm (b) - Edge 3-cycle clockwise  
./dist/cube lookup "u-perm (b)" --preview
# Algorithm: R2 U R U R' U' R' U' R' U R'

# H-Perm - Opposite edge swap
./dist/cube lookup "h-perm" --preview
# Algorithm: M2 U M2 U2 M2 U M2

# Z-Perm - Adjacent edge swap
./dist/cube lookup "z-perm" --preview
# Algorithm: M' U M2 U M2 U M' U2 M2
```

### 🧩 F2L (First Two Layers)

Build the first two layers efficiently with these fundamental patterns.

```bash
# Basic Insert - Standard corner-edge insertion
./dist/cube lookup "basic insert" --preview
# Algorithm: U R U' R'

# Split Pair - Advanced F2L technique
./dist/cube lookup "split pair" --preview  
# Algorithm: R U R' U' R U R' U'
```

### ⚡ Triggers (Fundamental Patterns)

The building blocks of speedcubing - learn these first!

```bash
# Sexy Move - Most important trigger
./dist/cube lookup "sexy move" --preview
# Algorithm: R U R' U'

# Sledgehammer - Essential F2L and OLL trigger
./dist/cube lookup "sledgehammer" --preview
# Algorithm: R' F R F'
```

## 🎨 Visual Previews

Use the `--preview` flag to see exactly what each algorithm does:

```bash
# See the cube state after applying Sune
./dist/cube lookup sune --preview

# Preview T-Perm with colors
./dist/cube lookup "t-perm" --preview --color

# Preview multiple algorithms
./dist/cube lookup --category OLL --preview
```

## 🔧 Advanced Search Features

### Case-Insensitive Search

```bash
# All of these work the same way
./dist/cube lookup SUNE
./dist/cube lookup sune  
./dist/cube lookup Sune
./dist/cube lookup SuNe
```

### Partial Name Matching

```bash
# Find algorithms with "perm" in the name
./dist/cube lookup perm

# Find algorithms with "u" in the name
./dist/cube lookup u

# Search for "oll" patterns
./dist/cube lookup oll
```

### Multiple Search Terms

```bash
# Search descriptions for specific terms
./dist/cube lookup corner
./dist/cube lookup edge
./dist/cube lookup cycle
./dist/cube lookup swap
```

## 🎯 Learning Workflows

### Complete Beginner Path

```bash
# 1. Start with triggers - the fundamentals
./dist/cube lookup "sexy move" --preview
./dist/cube lookup "sledgehammer" --preview

# 2. Learn basic F2L
./dist/cube lookup "basic insert" --preview

# 3. Master essential OLL cases
./dist/cube lookup sune --preview
./dist/cube lookup "anti-sune" --preview

# 4. Learn common PLL cases
./dist/cube lookup "t-perm" --preview
./dist/cube lookup "u-perm" --preview
```

### Category-Based Learning

```bash
# Master all triggers first
./dist/cube lookup --category Trigger

# Then move to F2L
./dist/cube lookup --category F2L --preview

# Learn OLL patterns
./dist/cube lookup --category OLL --preview

# Finally, conquer PLL
./dist/cube lookup --category PLL --preview
```

### Difficulty-Based Progression

```bash
# Easy (short algorithms)
./dist/cube lookup "sexy move"       # 4 moves
./dist/cube lookup "basic insert"    # 4 moves
./dist/cube lookup "sledgehammer"    # 4 moves

# Medium (moderate length)
./dist/cube lookup sune              # 7 moves
./dist/cube lookup "anti-sune"       # 7 moves
./dist/cube lookup "cross oll"       # 6 moves

# Advanced (longer algorithms)  
./dist/cube lookup "t-perm"          # 14 moves
./dist/cube lookup "y-perm"          # 17 moves
./dist/cube lookup "dot oll"         # 12 moves
```

## 🧪 Testing and Practice

### Verify Algorithm Execution

```bash
# Apply Sune and see the result
./dist/cube solve "R U R' U R U2 R'" --color

# Test T-Perm execution
./dist/cube solve "R U R' U' R' F R2 U' R' U' R U R' F'" --color

# Try the Sexy Move
./dist/cube solve "R U R' U'" --color
```

### Practice with Pattern Recognition

```bash
# Create a scramble, then apply an algorithm
./dist/cube solve "R U R' U'" --color                    # See the scramble
./dist/cube solve "R U R' U' U R U' R'" --color          # Apply Sune to it

# Study the pattern changes
./dist/cube show "R U R' U'" --highlight-oll --color     # Before
./dist/cube show "R U R' U' R U R' U R U2 R'" --highlight-oll --color  # After
```

### Create Custom Training Sets

```bash
# Practice all OLL cases in sequence
for alg in sune "anti-sune" "cross oll" "dot oll" "l-shape"; do
    echo "=== $alg ==="
    ./dist/cube lookup "$alg" --preview
    echo ""
done

# Practice all PLL cases
for alg in "t-perm" "y-perm" "u-perm (a)" "u-perm (b)" "h-perm" "z-perm"; do
    echo "=== $alg ==="
    ./dist/cube lookup "$alg" --preview  
    echo ""
done
```

## 🔍 Algorithm Discovery

Use the `cube find` command to discover your own algorithms or find alternatives:

### Find Short Solutions

```bash
# Find short ways to achieve common patterns
./dist/cube find pattern solved --max-moves 4
./dist/cube find pattern cross --max-moves 6
```

### Discover Alternative Algorithms

```bash
# Find different ways to create the same result as Sune
./dist/cube find sequence "R U R' U R U2 R'" --max-moves 8

# Discover alternatives to T-Perm
./dist/cube find sequence "R U R' U' R' F R2 U' R' U' R U R' F'" --max-moves 10
```

### Create Custom Algorithms

```bash
# Start from a specific position and find solutions
./dist/cube find pattern solved --from "R U R' U'" --max-moves 6

# Discover algorithms from complex starting positions
./dist/cube find pattern cross --from "F R U' R' F'" --max-moves 8
```

## 🎯 Expert Tips

### Memory Aids

```bash
# Use descriptive searches to remember algorithms
./dist/cube lookup "corner swap"     # Finds PLL algorithms
./dist/cube lookup "edge cycle"      # Finds U-Perms
./dist/cube lookup "cross"           # Finds cross-related algorithms
./dist/cube lookup "trigger"         # Finds fundamental patterns
```

### Algorithm Analysis

```bash
# Study algorithm structure by searching for specific moves
./dist/cube lookup --pattern "R U R'"    # Find trigger-based algorithms
./dist/cube lookup --pattern "M2"        # Find M-slice algorithms
./dist/cube lookup --pattern "F'"        # Find algorithms using F'
```

### Performance Optimization

```bash
# Find the shortest algorithms in each category
./dist/cube lookup --category Trigger     # Shortest: 4 moves
./dist/cube lookup --category F2L         # Shortest: 4 moves  
./dist/cube lookup --category OLL         # Shortest: 6 moves
./dist/cube lookup --category PLL         # Shortest: 7 moves
```

## 📊 Complete Algorithm Reference

### Quick Reference Table

| Algorithm | Category | Moves | Length |
|-----------|----------|-------|---------|
| Sexy Move | Trigger | `R U R' U'` | 4 |
| Sledgehammer | Trigger | `R' F R F'` | 4 |
| Basic Insert | F2L | `U R U' R'` | 4 |
| Cross OLL | OLL | `F R U R' U' F'` | 6 |
| Sune | OLL | `R U R' U R U2 R'` | 7 |
| Anti-Sune | OLL | `R U2 R' U' R U' R'` | 7 |
| H-Perm | PLL | `M2 U M2 U2 M2 U M2` | 7 |
| Split Pair | F2L | `R U R' U' R U R' U'` | 8 |
| L-Shape OLL | OLL | `F U R U' R' F'` | 6 |
| Z-Perm | PLL | `M' U M2 U M2 U M' U2 M2` | 9 |
| U-Perm (a) | PLL | `R U' R U R U R U' R' U' R2` | 11 |
| U-Perm (b) | PLL | `R2 U R U R' U' R' U' R' U R'` | 11 |
| Dot OLL | OLL | `F R U R' U' F' f R U R' U' f'` | 12 |
| T-Perm | PLL | `R U R' U' R' F R2 U' R' U' R U R' F'` | 14 |
| Y-Perm | PLL | `F R U' R' U' R U R' F' R U R' U' R' F R F'` | 17 |

### Search Examples for Each Algorithm

```bash
# By exact name
./dist/cube lookup "Sexy Move"
./dist/cube lookup "T-Perm"  
./dist/cube lookup "Anti-Sune"

# By case number
./dist/cube lookup "OLL 27"    # Sune
./dist/cube lookup "PLL T"     # T-Perm
./dist/cube lookup "F2L 1"     # Basic Insert

# By moves pattern
./dist/cube lookup --pattern "R U R' U'"     # Sexy Move and others
./dist/cube lookup --pattern "M2"            # H-Perm and Z-Perm
./dist/cube lookup --pattern "F R U R' U' F'" # Cross OLL and Dot OLL
```

## 🎯 Next Steps

Ready to put your algorithm knowledge to work?

1. **Practice with real solving**: Use algorithms in the [Basic Solving Guide](./solving.md)
2. **Apply advanced techniques**: Combine with [Advanced Features](./advanced.md)
3. **Try the web interface**: Access algorithms via [Web Interface](./web.md)

**Master these 15 algorithms and you'll have a solid foundation for speedcubing success!** 🎲✨