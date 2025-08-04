# Algorithm Database Refactor Proposal

## Current Issues

### 1. Redundant Verification System
The current `Verified` boolean and `TestedOn` array are conceptually flawed:
- If an algorithm is in the database with patterns, it's verified by definition
- Algorithms that work on 3x3 inherently work on larger cubes (outer layer moves)
- The `TestedOn` array suggests algorithms might work differently on different cube sizes, which is incorrect

### 2. Inflexible Pattern Representation
The current `StartCFEN`/`TargetCFEN` approach:
- Assumes specific starting states (often "solved")
- Doesn't clearly show what the algorithm actually changes
- Makes it hard to understand algorithm effects at a glance
- Requires full CFEN strings even when most stickers don't change

### 3. Missing Key Metadata
The current structure lacks:
- Clear indication of which stickers are affected
- Standard notation variants (e.g., different fingertrick-friendly versions)
- Recognition patterns (what to look for to know when to use this algorithm)
- Inverse algorithms (important for drilling and practice)

## Proposed Structure

```go
type Algorithm struct {
    // Core Identity
    Name        string   // e.g., "Sune"
    CaseID      string   // e.g., "OLL-27" (standardized format)
    Category    string   // OLL, PLL, F2L, Trigger, etc.
    
    // Algorithm Definition
    Moves       string   // e.g., "R U R' U R U2 R'"
    MoveCount   int      // Auto-calculated from Moves
    
    // Pattern Representation (NEW APPROACH)
    Pattern     string   // Masked CFEN showing only affected stickers
    // Example: "YB|GY5***/*G2*6/*Y*Y*6/--/***O**/***G**"
    // Where: * = grey (unchanged), actual colors = changed stickers
    
    // Human-Friendly Info
    Description string   // What this algorithm does
    Recognition string   // How to recognize when to use it
    
    // Optional Metadata
    Probability float64  // Chance of occurring in solve
    Variants    []string // Alternative move sequences
    Inverse     string   // Inverse algorithm (if meaningful)
    
    // Relationships
    Mirror      string   // ID of mirror algorithm (e.g., "OLL-26" for Sune)
    Related     []string // IDs of related algorithms
}
```

## Pattern Generation Process

Starting from a solved cube with canonical YB orientation:
1. Apply the algorithm
2. Compare before/after states
3. Create masked CFEN where:
   - Stickers that changed: Show new color
   - Stickers that stayed same: Replace with `*` (grey/wildcard)
   - This clearly shows the algorithm's effect

Example for Sune (R U R' U R U2 R'):
```
Before: YB|Y9/R9/B9/W9/O9/G9
After:  YB|BY5RYG/YO2R6/YBOB6/W9/YG2O6/BR2G6
Masked: YB|*Y5*Y*/*O2*6/Y*O*6/*9/Y*2O6/**2*6
```

## Benefits of New Structure

### 1. Clarity
- Pattern immediately shows what changes
- No confusion about "verification" - if it has a pattern, it's verified
- Clear standardized case IDs (OLL-27 vs "OLL 27" inconsistency)

### 2. Flexibility
- Patterns work regardless of starting state (as long as affected pieces are solved)
- Easy to see algorithm relationships (similar patterns)
- Can identify "inverse pairs" automatically

### 3. Extensibility
- Easy to add new algorithms with consistent format
- Pattern format supports partial matching for advanced features
- Could later add animation data, finger tricks, etc.

## Migration Strategy

### Phase 1: Update Structure
```go
// Keep old structure temporarily for migration
type AlgorithmV2 struct {
    Algorithm  // New structure
    LegacyData *Algorithm // Old structure for reference
}
```

### Phase 2: Pattern Generation Tool
Create a tool that:
1. Takes algorithm moves
2. Applies to solved cube
3. Generates masked pattern automatically
4. Validates pattern is minimal (no unnecessary color info)

### Phase 3: Bulk Migration
1. Auto-generate patterns for all algorithms
2. Standardize case IDs (OLL-1 through OLL-57, etc.)
3. Add recognition descriptions
4. Identify mirror/inverse relationships

### Phase 4: Cleanup
1. Remove legacy fields
2. Update all code to use new structure
3. Update lookup/search functions

## Example Migrated Entries

```go
{
    Name:        "Sune",
    CaseID:      "OLL-27",
    Category:    "OLL",
    Moves:       "R U R' U R U2 R'",
    MoveCount:   7,
    Pattern:     "YB|*Y5*Y*/*O2*6/Y*O*6/*9/Y*2O6/**2*6",
    Description: "Orients corners when one is correctly oriented",
    Recognition: "One corner oriented, headlights on left",
    Probability: 4.63,  // 1/216 * 1000
    Inverse:     "R U2 R' U' R U' R'",  // This is Anti-Sune
    Mirror:      "OLL-26",  // Anti-Sune
    Related:     []string{"OLL-26", "OLL-21"},  // Anti-Sune, Double Sune
},
{
    Name:        "T-Perm",
    CaseID:      "PLL-T",
    Category:    "PLL",
    Moves:       "R U R' U' R' F R2 U' R' U' R U R' F'",
    MoveCount:   14,
    Pattern:     "YB|*9/*O**6/*2**6/*9/O*O7/*G8",
    Description: "Swaps two adjacent corners and two edges",
    Recognition: "Headlights with opposite edge swap",
    Probability: 4.17,  // 1/72 * 3
    Variants:    []string{"R U R' U' R' F R2 U' R' U' R U R' F'"},
},
{
    Name:        "Sexy Move",
    CaseID:      "TRIG-1",
    Category:    "Trigger",
    Moves:       "R U R' U'",
    MoveCount:   4,
    Pattern:     "YB|**Y6*/**2*6/*5Y*2Y/W2*W6/YO8/O**7",
    Description: "Most common trigger in cubing",
    Recognition: "F2L pair building/breaking trigger",
    Related:     []string{"TRIG-2", "TRIG-3"},  // Sledgehammer, Lefty Sexy
}
```

## Implementation Priority

1. **Must Have**
   - New Algorithm struct
   - Pattern generation tool
   - Migration of existing 67 algorithms

2. **Should Have**
   - Recognition descriptions for all cases
   - Standardized case IDs
   - Inverse algorithm identification

3. **Nice to Have**
   - Probability calculations
   - Animation preview data
   - Fingertrick annotations

## Conclusion

This refactor simplifies the algorithm database while making it more powerful and maintainable. The masked pattern approach elegantly captures what each algorithm does without overspecifying requirements. By removing redundant concepts like "verified" and "testedOn", we focus on what matters: the algorithm's effect on the cube.