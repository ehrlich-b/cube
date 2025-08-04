# Move Visualization Design

## Current State

Currently, the cube CLI displays the entire cube in an unfolded cross pattern. This works well for general visualization but is inefficient for last layer algorithms where most of the cube doesn't change.

## Proposed Enhancement: Last Layer Mode

When an algorithm only affects the top two layers (detected via pattern masks), switch to a focused view showing just the top layer surrounded by the edge stickers:

```
Full cube view (current):           Last Layer view (proposed):
    YBY                                    🟩 🟧 🟩
    YYY                               🟧 🟨🟨🟨 🟥
    YYY                               🟧 🟨🟨🟨 🟥
                                      🟧 🟨🟨🟨 🟥
YRR WWG OOB YOO                            🟦🟦🟦
RRR WWB YOO YYY
RRR WWW BOO YYY

    GGO
    GGG
    GGG
```

## Implementation

### 1. Detection
```go
func AffectsOnlyLastLayers(pattern string) bool {
    // Check if pattern only has non-* in top 2 layers
}
```

### 2. Rendering
```go
func RenderLastLayer(cube *Cube, useColor bool) string {
    // Layout:
    //     [back edge row]
    // [left] [top face] [right]
    //     [front edge row]
}
```

### 3. Display Logic
- **OLL/PLL algorithms**: Use last layer view
- **F2L algorithms**: Use full cube view
- **Other algorithms**: Auto-detect based on pattern

## CLI Usage
```bash
cube show-alg "R U R' U R U2 R'"              # Auto-detect
cube show-alg "R U R' U R U2 R'" --view=last  # Force last layer
cube show-alg "R U R' U R U2 R'" --view=full  # Force full cube
```

## Benefits
- Focus on what changes
- Less vertical space
- Natural for OLL/PLL recognition
- Simple implementation using existing emoji blocks