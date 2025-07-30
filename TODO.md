# TODO.md - Rubik's Cube Solver Development Plan

## 🚨 PHASE 1: Make It Actually Work (Core Fixes)

**Goal**: Transform placeholder code into a working cube solver with beautiful output

### Critical Fixes
- [ ] **Fix the broken move system** (`internal/cube/moves.go:112`)
  - Currently only rotates face matrix, doesn't move adjacent edges
  - Implement proper edge/corner movement for 3x3 cube first
  - Test with known scrambles: `R U R' U'` should actually scramble the cube
  
- [ ] **Replace one fake solver with real logic**
  - Pick BeginnerSolver and implement basic layer-by-layer
  - Just needs: white cross → white corners → middle layer → yellow cross → yellow corners
  - Don't overthink it - basic working algorithm is the goal

- [ ] **Add beautiful ASCII color output**
  - Color cube faces with ANSI colors or Unicode squares (🟨🟩🟦🟥🟧⬜)
  - Make cube visualization actually look good in terminal
  - Add `--color` flag to CLI commands

### Success Criteria for Phase 1
- `cube solve "R U R' U'"` produces different output than solved cube
- BeginnerSolver actually solves a scrambled 3x3 cube
- ASCII output looks professional with colors
- All existing tests pass (add basic tests if none exist)

**⚠️ Before Phase 2: Re-read this TODO.md, examine what you built, and adjust the plan**

---

## 🎯 PHASE 2: Essential Cuber Features

**Goal**: Add the features that make this tool actually useful for cubers

### Core Functionality
- [ ] **"Is this solved?" verification**
  - `cube verify "scramble" "solution"` command
  - Apply scramble, then solution, check if solved
  
- [ ] **Pattern highlighting system**
  - Grey out squares that don't matter for specific algorithms
  - `cube show --highlight-cross` or `cube show --highlight-oll`
  
- [ ] **Algorithm database (start small)**
  - Just store 10-15 common OLL/PLL cases with names
  - `cube lookup --pattern "R U R' U'"` shows algorithm name/description
  - Basic pattern matching against current cube state

### Terminal Interface Improvements
- [ ] **Interactive cube viewer**
  - `cube interactive` - apply moves with keyboard, see results immediately
  - Arrow keys navigate, spacebar applies move, 'q' quits
  
- [ ] **Better move notation support**
  - Wide turns: `Rw`, slice moves: `M E S`  
  - Rotations: `x y z`
  - Parse and display properly

### Success Criteria for Phase 2
- Can verify if a solution actually solves a scramble
- Can highlight specific patterns on cube display
- Interactive mode lets you "play" with the cube
- Small but useful algorithm database working

**⚠️ Before Phase 3: Re-read this TODO.md, test everything thoroughly, adjust plan based on what you learned**

---

## 🌐 PHASE 3: Terminal Web Interface 

**Goal**: Create web interface that feels like using the terminal

### Web Terminal Implementation
- [ ] **Mirror CLI functionality in web**
  - Web interface calls same backend functions as CLI
  - POST `/api/exec` endpoint that runs CLI commands
  - Return ANSI colored output directly to web terminal

- [ ] **Terminal emulator styling**
  - Monospace font, dark theme, cursor prompt
  - Command history with arrow keys
  - Copy-paste friendly output

- [ ] **Real-time features** 
  - WebSocket for interactive commands
  - Streaming output for long operations
  - Session persistence

### Success Criteria for Phase 3
- Web interface looks and feels like terminal
- All CLI commands work identically in web browser
- Can share cube states via URLs
- Mobile-friendly terminal interface

**⚠️ Before Phase 4: Re-evaluate entire project, check if users actually want these features**

---

## 🔥 PHASE 4: Power User Tools

**Goal**: Features that serious cubers would find genuinely useful

### Algorithm Discovery
- [ ] **Simple exhaustive search**
  - `cube find --pattern "cross" --max-moves 8`
  - Find sequences that create specific patterns
  
- [ ] **Algorithm trainer**
  - Show scrambled cube, user inputs solution
  - Verify correctness and timing
  
- [ ] **Pattern analysis**
  - Identify what patterns exist on current cube
  - Suggest relevant algorithms

### Advanced Features
- [ ] **Scramble generation**
  - Generate random state scrambles
  - Generate scrambles for specific practice (cross, OLL cases)

- [ ] **Move optimization**
  - Simplify move sequences: `R R` → `R2`, `R R'` → remove

### Success Criteria for Phase 4
- Can discover new algorithms through search
- Training mode helps learn existing algorithms
- Advanced users find it genuinely useful for practice

**⚠️ Before Phase 5: Seriously evaluate if more features are needed or if you should polish what exists**

---

## 🚀 PHASE 5: Polish & Performance

**Goal**: Make it production-ready and fast

### Performance & Reliability
- [ ] **Comprehensive test suite**
  - Unit tests for all core functions
  - Integration tests for CLI commands
  - Algorithm correctness verification

- [ ] **Performance optimization**
  - Fast cube state representation
  - Efficient algorithm search
  - Benchmark and optimize hot paths

### User Experience
- [ ] **Error handling & help**
  - Clear error messages for invalid moves
  - Built-in help system and examples
  - Onboarding for new users

### Success Criteria for Phase 5
- No bugs in core functionality
- Fast enough for real-time use
- Professional-quality user experience

---

## 📋 Implementation Notes

### Between Each Phase:
1. **Stop and re-read this TODO.md**
2. **Test everything you've built so far**
3. **Ask: "Is this still the right direction?"**
4. **Adjust the plan based on what you learned**
5. **Don't over-engineer - ship working features first**

### Key Principles:
- **Start simple**: Get basic functionality working before adding complexity
- **Test everything**: Don't move to next phase with broken features
- **User feedback**: Build features cubers actually want, not just cool tech
- **Iterative improvement**: Each phase should leave you with a usable tool

### Current Blockers:
- Move system is fundamentally broken (Phase 1 priority)
- All solvers are fake (Phase 1 priority)  
- No visual appeal (Phase 1 priority)

**Remember**: A simple tool that works is infinitely better than a complex tool that doesn't.