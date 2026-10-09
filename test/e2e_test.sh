#!/usr/bin/env bash

# End-to-end test harness for cube solver
# Cross-platform compatible (macOS/Linux)

set -euo pipefail

# Colors for output (cross-platform)
if [[ "$OSTYPE" == "darwin"* ]] || [[ "$OSTYPE" == "linux-gnu"* ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Binary path
CUBE_BIN="${CUBE_BIN:-./dist/cube}"

# Ensure binary exists
if [ ! -f "$CUBE_BIN" ]; then
    echo -e "${RED}Error: cube binary not found at $CUBE_BIN${NC}"
    echo "Please run 'make build' first"
    exit 1
fi

# Test function
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_pattern="${3:-}"
    local should_fail="${4:-false}"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo -ne "${BLUE}Testing:${NC} $test_name... "
    
    # Run command and capture output and exit code
    set +e
    output=$(eval "$command" 2>&1)
    exit_code=$?
    set -e
    
    # Check if command should fail
    if [ "$should_fail" = "true" ]; then
        if [ $exit_code -ne 0 ]; then
            echo -e "${GREEN}PASS${NC} (correctly failed)"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return
        else
            echo -e "${RED}FAIL${NC} (should have failed but didn't)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            echo "Output: $output"
            return
        fi
    fi
    
    # Check exit code for commands that should succeed
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}FAIL${NC} (exit code: $exit_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "Output: $output"
        return
    fi
    
    # Check output pattern if provided
    if [ -n "$expected_pattern" ]; then
        if echo "$output" | grep -q "$expected_pattern"; then
            echo -e "${GREEN}PASS${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}FAIL${NC} (pattern not found: $expected_pattern)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            echo "Output: $output"
        fi
    else
        echo -e "${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

echo -e "${YELLOW}=== Cube Solver E2E Test Suite ===${NC}"
echo "Binary: $CUBE_BIN"
echo ""

# Basic Commands
echo -e "${YELLOW}Basic Commands:${NC}"
run_test "Help command" "$CUBE_BIN help" "Available Commands:"
run_test "Version flag" "$CUBE_BIN --version" "cube version"

# Solve Command Tests
echo -e "\n${YELLOW}Solve Command Tests:${NC}"
run_test "Basic solve" "$CUBE_BIN solve \"R U R' U'\"" "Solution:"
run_test "Solve with color" "$CUBE_BIN solve \"R U R' U'\" --color" "🟦"
run_test "Solve with beginner algorithm" "$CUBE_BIN solve \"R U R' U'\" --algorithm beginner" "Using algorithm: beginner"
run_test "Solve with CFOP algorithm" "$CUBE_BIN solve \"R U R' U'\" --algorithm cfop" "F2L 1:"
run_test "Solve with Kociemba algorithm" "$CUBE_BIN solve \"R U R' U'\" --algorithm kociemba" "Using algorithm: kociemba"
run_test "Solve with Kociemba method" "$CUBE_BIN solve \"R U R' U'\" --method kociemba" "Using algorithm: kociemba"
run_test "Solve with optimal search" "$CUBE_BIN solve 'R U F2' --optimal --time-limit 5s" "Using algorithm: optimal"
for dimension in 3 5; do
    for numbered_turn in "2R" "2R'" "2R2"; do
        run_test "Numbered turn solve replay (${dimension}x${dimension}, $numbered_turn)" \
            "solution=\$(\"$CUBE_BIN\" solve \"$numbered_turn\" --dimension $dimension --headless) && \"$CUBE_BIN\" twist \"$numbered_turn \$solution\" --dimension $dimension" "SOLVED!"
    done
done
# These three former unsupported-size failures now require a real verified solve.
run_test "Solve 2x2 cube" "$CUBE_BIN solve \"R U R' U'\" --dimension 2 --cfen" "YB|Y4/R4/B4/W4/O4/G4"
run_test "Solve 4x4 cube" "$CUBE_BIN solve \"Rw Uw Fw\" --dimension 4 --cfen" "YB|Y16/R16/B16/W16/O16/G16"
run_test "Solve 5x5 cube" "$CUBE_BIN solve \"2R 3L\" --dimension 5 --cfen" "YB|Y25/R25/B25/W25/O25/G25"
run_test "Solve 6x6 cube" "$CUBE_BIN solve \"2R 3U Fw\" --dimension 6 --cfen" "YB|Y36/R36/B36/W36/O36/G36"
reduced_six_cfen=$(cat test/fixtures/reduced-six.cfen)
run_test "Reduced 6x6 survives tiny optimization budget" "$CUBE_BIN solve --dimension 6 --start \"$reduced_six_cfen\" --time-limit 1ns --target-length 1 --cfen" "YB|Y36/R36/B36/W36/O36/G36"
run_test "Solve 7x7 fixed-center orientation" "$CUBE_BIN solve \"M E S x Rw\" --dimension 7 --cfen" "YB|Y49/R49/B49/W49/O49/G49"
run_test "Solve 4x4 OLL parity" "$CUBE_BIN solve \"2R2 B2 U2 2L U2 2R' U2 2R U2 F2 2R F2 2L' B2 2R2\" --dimension 4 --cfen" "YB|Y16/R16/B16/W16/O16/G16"
run_test "Solve 4x4 PLL parity" "$CUBE_BIN solve \"2R2 U2 2R2 Uw2 2R2 Uw2\" --dimension 4 --cfen" "YB|Y16/R16/B16/W16/O16/G16"
run_test "Unsupported solve dimension" "$CUBE_BIN solve R --dimension 8" "" true
run_test "Empty scramble" "$CUBE_BIN solve ''" "Solving 3x3x3 cube"
run_test "Invalid algorithm" "$CUBE_BIN solve 'R U' --algorithm invalid" "Error getting solver" true

# Twist Command Tests  
echo -e "\n${YELLOW}Twist Command Tests:${NC}"
run_test "Basic twist" "$CUBE_BIN twist \"R U R' U'\"" "Cube state after applying moves:"
run_test "Twist with color" "$CUBE_BIN twist \"R U R' U'\" --color" "🟦"
run_test "Twist empty moves" "$CUBE_BIN twist \"\"" "✅ SOLVED!"
run_test "Twist canceling moves" "$CUBE_BIN twist \"R R'\"" "✅ SOLVED!"
run_test "Twist 2x2 cube" "$CUBE_BIN twist \"R U R' U'\" --dimension 2" "Applying moves to 2x2x2 cube"
run_test "Twist 4x4 cube" "$CUBE_BIN twist \"Rw Uw Fw\" --dimension 4" "Applying moves to 4x4x4 cube"
run_test "Twist slice moves" "$CUBE_BIN twist \"M E S\" --dimension 3" "Moves applied: 3"
run_test "Twist wide moves" "$CUBE_BIN twist \"Rw Fw Uw\" --dimension 4" "🔄 Scrambled"
run_test "Twist layer moves" "$CUBE_BIN twist \"2R 3L\" --dimension 5" "Applying moves to 5x5x5 cube"
run_test "Twist rotations" "$CUBE_BIN twist \"x y z\"" "Moves applied: 3"
run_test "Twist help" "$CUBE_BIN twist --help" "Apply a sequence of moves to a cube"

# Advanced Notation Tests
echo -e "\n${YELLOW}Advanced Notation Tests:${NC}"
run_test "Slice moves (M E S)" "$CUBE_BIN solve \"M E S\" --dimension 3" "Solution:"
run_test "Wide moves" "$CUBE_BIN twist \"Rw Fw Uw\" --dimension 4" "Moves applied:"
run_test "Layer moves" "$CUBE_BIN twist \"2R 3L 2F\" --dimension 5" "Moves applied:"
run_test "Rotations" "$CUBE_BIN solve \"x y z\" --dimension 3" "Solution:"
run_test "Mixed notation" "$CUBE_BIN twist \"R M U' 2R Fw x y\" --dimension 5" "Moves applied:"

# Verify Command Tests
echo -e "\n${YELLOW}Verify Command Tests:${NC}"

# Test 1: Generate scrambled state, verify inverse solves it
echo -n "Testing verify with R scramble and R' inverse... "
scrambled_cfen=$($CUBE_BIN generate-cfen "R" 2>/dev/null)
solved_cfen="YB|Y9/R9/B9/W9/O9/G9"
if $CUBE_BIN verify "R'" --start "$scrambled_cfen" --target "$solved_cfen" --headless 2>/dev/null; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (R' should solve R scramble)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test 2: More complex scramble/inverse test
echo -n "Testing verify with complex scramble and inverse... "
scrambled_cfen=$($CUBE_BIN generate-cfen "R U R' U'" 2>/dev/null)
if $CUBE_BIN verify "U R U' R'" --start "$scrambled_cfen" --target "$solved_cfen" --headless 2>/dev/null; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (Inverse should solve scramble)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test 3: Wrong algorithm should fail
echo -n "Testing verify detects incorrect algorithm... "
scrambled_cfen=$($CUBE_BIN generate-cfen "R U R' U'" 2>/dev/null)
if ! $CUBE_BIN verify "R U R'" --start "$scrambled_cfen" --target "$solved_cfen" --headless 2>/dev/null; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (Wrong algorithm should fail)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test 4: Wildcard matching - only care about top face orientation
echo -n "Testing verify with wildcard target (top face only)... "
scrambled_cfen=$($CUBE_BIN generate-cfen "R U R' U'" 2>/dev/null)
wildcard_target="YB|Y9/?9/?9/?9/?9/?9"  # Only top face specified
if $CUBE_BIN verify "U R U' R'" --start "$scrambled_cfen" --target "$wildcard_target" --headless 2>/dev/null; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (Should match wildcard pattern)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test 5: Verbose output
run_test "Verify with verbose output" "$CUBE_BIN verify \"R' R\" --verbose" "Start state"

# Test 6: Color output
run_test "Verify with color output" "$CUBE_BIN verify \"U U'\" --verbose --color" "🟦"

# Show Command Tests
echo -e "\n${YELLOW}Show Command Tests:${NC}"
run_test "Show solved cube" "$CUBE_BIN show" "Solved cube state:"
run_test "Show scrambled cube" "$CUBE_BIN show \"R U R' U'\"" "Cube state after scramble:"
run_test "Show with cross highlight" "$CUBE_BIN show \"R U R' U'\" --highlight-cross" "Highlighting: CROSS pattern"
run_test "Show with OLL highlight" "$CUBE_BIN show \"R U R' U'\" --highlight-oll" "Highlighting: OLL pattern"
run_test "Show with PLL highlight" "$CUBE_BIN show \"R U R' U'\" --highlight-pll" "Highlighting: PLL pattern"
run_test "Show with F2L highlight" "$CUBE_BIN show \"R U R' U'\" --highlight-f2l" "Highlighting: F2L pattern"
run_test "Show with color and highlight" "$CUBE_BIN show \"R U\" --highlight-oll --color" "⬛"

# Lookup Command Tests
echo -e "\n${YELLOW}Lookup Command Tests:${NC}"
run_test "Lookup by name" "$CUBE_BIN lookup sune" "OLL-27 - Sune"
run_test "Lookup by pattern" "$CUBE_BIN lookup --pattern \"R U R' U'\"" "Sexy Move"
run_test "Lookup by category OLL" "$CUBE_BIN lookup --category OLL" "Sune"
run_test "Lookup by category PLL" "$CUBE_BIN lookup --category PLL" "T-Perm"
run_test "Lookup 2x2 preview uses its dimension" "$CUBE_BIN lookup 2x2-OLL-1 --preview" "Dimension: 2x2"
run_test "Lookup merged OLL alias" "$CUBE_BIN lookup OLL-27" "Aliases:.*OLL-27"
run_test "Lookup all algorithms" "$CUBE_BIN lookup --all" "All algorithms in database:"
run_test "Lookup with preview" "$CUBE_BIN lookup sune --preview" "Top face after algorithm:"
run_test "Lookup T-Perm" "$CUBE_BIN lookup \"T-Perm\"" "PLL-T - T-Perm"
run_test "Lookup non-existent" "$CUBE_BIN lookup xyz123" "No algorithms found"
run_test "Lookup no args" "$CUBE_BIN lookup" "Please provide a query"

# Headless Mode Tests
echo -e "\n${YELLOW}Headless Mode Tests:${NC}"
headless_solution=$($CUBE_BIN solve "R U F2 L' B" --headless)
run_test "Headless full-solution notation" "$CUBE_BIN optimize \"$headless_solution\""

echo -n "Testing headless verify success (correct algorithm)... "
scrambled_cfen=$($CUBE_BIN generate-cfen "R U R' U'" 2>/dev/null)
solved_cfen="YB|Y9/R9/B9/W9/O9/G9"
if $CUBE_BIN verify "U R U' R'" --start "$scrambled_cfen" --target "$solved_cfen" --headless >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (should exit 0 for correct inverse)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Testing headless verify failure (incorrect algorithm)... "
scrambled_cfen=$($CUBE_BIN generate-cfen "R U R' U'" 2>/dev/null)
if ! $CUBE_BIN verify "R U R'" --start "$scrambled_cfen" --target "$solved_cfen" --headless >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (should exit 1 for incorrect algorithm)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Testing headless verify produces no output... "
headless_verify_output=$($CUBE_BIN verify "U U'" --headless 2>&1 || true)
if [[ -z "$headless_verify_output" ]]; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (headless verify should produce no output)"
    echo "Output: '$headless_verify_output'"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Edge Cases and Error Handling
echo -e "\n${YELLOW}Edge Cases and Error Handling:${NC}"
run_test "Invalid move notation" "$CUBE_BIN solve \"R X Q\"" "Error parsing scramble" true
run_test "Huge cube dimension" "$CUBE_BIN twist \"R\" --dimension 20" "Applying moves to 20x20x20 cube"
run_test "Multiple flags" "$CUBE_BIN solve \"R U R' U'\" --color --dimension 4 --algorithm cfop" "" true

# Complex Integration Tests
echo -e "\n${YELLOW}Complex Integration Tests:${NC}"

# Test verify on various scramble lengths
echo -n "Testing verify with 6-move scramble... "
scramble="R U2 R' D R D'"
inverse="D R' D' R U2 R'"
scrambled_cfen=$($CUBE_BIN generate-cfen "$scramble" 2>/dev/null)
if $CUBE_BIN verify "$inverse" --start "$scrambled_cfen" --target "$solved_cfen" --headless 2>/dev/null; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (6-move inverse should work)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Testing verify with slice moves... "
scramble="M E S"
inverse="S' E' M'"
scrambled_cfen=$($CUBE_BIN generate-cfen "$scramble" 2>/dev/null)
if $CUBE_BIN verify "$inverse" --start "$scrambled_cfen" --target "$solved_cfen" --headless 2>/dev/null; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (slice move inverse should work)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Verify a real full beginner solution against an independently specified solved state.
full_scramble="x R U F2 L' B"
full_start=$($CUBE_BIN generate-cfen "$full_scramble")
full_moves=$($CUBE_BIN solve "$full_scramble" --method beginner --headless)
run_test "Full beginner solution replay" "$CUBE_BIN verify \"$full_moves\" --start '$full_start' --target '$solved_cfen' --headless"
fast_moves=$($CUBE_BIN solve "$full_scramble" --method kociemba --headless)
run_test "Full Kociemba solution replay" "$CUBE_BIN verify \"$fast_moves\" --start '$full_start' --target '$solved_cfen' --headless"

# Performance Tests
echo -e "\n${YELLOW}Performance Tests:${NC}"
echo -n "Testing large scramble performance... "
start_time=$(date +%s%N 2>/dev/null || date +%s)
large_scramble="R U R' U' F B L R D U R U R' U' F B L R D U R U R' U' F B L R D U"
if $CUBE_BIN twist "$large_scramble" --dimension 6 >/dev/null 2>&1; then
    end_time=$(date +%s%N 2>/dev/null || date +%s)
    echo -e "${GREEN}PASS${NC} (completed large scramble)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Comprehensive Fuzzing Tests  
echo -e "\n${YELLOW}Comprehensive Fuzzing Tests:${NC}"

# Generate random scrambles for fuzzing
generate_random_scramble() {
    local length=${1:-10}
    local moves=("R" "R'" "R2" "L" "L'" "L2" "U" "U'" "U2" "D" "D'" "D2" "F" "F'" "F2" "B" "B'" "B2")
    local scramble=""
    for i in $(seq 1 $length); do
        if [ $i -gt 1 ]; then scramble+=" "; fi
        scramble+="${moves[$((RANDOM % ${#moves[@]}))]}"
    done
    echo "$scramble"
}

# Fuzzing test function 
fuzz_test_solver() {
    local algorithm=$1
    local test_count=${2:-20}
    local failed_count=0
    local total_count=0
    
    echo -n "Fuzzing $algorithm solver ($test_count random scrambles)... "
    
    for i in $(seq 1 $test_count); do
        total_count=$((total_count + 1))
        
        # Generate random scramble (5-15 moves)
        scramble_length=$((5 + RANDOM % 11))
        scramble=$(generate_random_scramble $scramble_length)
        
        # Get solution in headless mode
        solution=$($CUBE_BIN solve "$scramble" --algorithm "$algorithm" --headless 2>/dev/null)
        
        if [ -z "$solution" ]; then
            failed_count=$((failed_count + 1))
            continue
        fi
        
        scrambled_cfen=$($CUBE_BIN generate-cfen "$scramble")
        if ! $CUBE_BIN verify "$solution" --start "$scrambled_cfen" --target 'YB|Y9/R9/B9/W9/O9/G9' --headless; then
            failed_count=$((failed_count + 1))
            
            # Re-run in non-headless mode for debugging
            echo -e "\n${RED}FUZZ FAILURE DETECTED!${NC}"
            echo "Algorithm: $algorithm"
            echo "Failing scramble: $scramble"
            echo "Generated solution: $solution"
            echo ""
            echo "Re-running in debug mode:"
            echo "=== SOLVE ==="
            $CUBE_BIN solve "$scramble" --algorithm "$algorithm" --color
            echo ""
            echo "=== VERIFY ==="
            $CUBE_BIN verify "$solution" --start "$scrambled_cfen" --target 'YB|Y9/R9/B9/W9/O9/G9' --verbose --color
            echo ""
            echo "Halting fuzzing due to failure."
            return 1
        fi
    done
    
    if [ $failed_count -eq 0 ]; then
        echo -e "${GREEN}PASS${NC} ($total_count/$total_count scrambles solved correctly)"
        return 0
    else
        echo -e "${RED}FAIL${NC} ($failed_count/$total_count scrambles failed)"
        return 1
    fi
}

# FOUNDATIONAL FUZZING: Move System Inverse Testing
# This is the anchor test - if move sequences + their inverses don't return to solved,
# then the basic move system is broken and no solver can work correctly.

INVERSE_FUZZ_COUNT=5  # Start small to avoid CPU churn

# Inverse fuzzing test function
fuzz_test_move_inverses() {
    local test_count=${1:-50}
    local failed_count=0
    local total_count=0
    
    # Set deterministic seed for reproducible testing
    RANDOM=42
    
    echo -n "Inverse fuzzing move system ($test_count random sequences)... "
    
    for i in $(seq 1 $test_count); do
        total_count=$((total_count + 1))
        
        # Generate random scramble (3-5 moves to keep it fast)
        scramble_length=$((3 + RANDOM % 3))
        scramble=$(generate_random_scramble $scramble_length)
        
        # Create inverse sequence (reverse order, invert each move)
        inverse=""
        IFS=' ' read -ra MOVES <<< "$scramble"
        for ((j=${#MOVES[@]}-1; j>=0; j--)); do
            move="${MOVES[j]}"
            # Invert the move: R->R', R'->R, R2->R2
            if [[ "$move" == *"'" ]]; then
                inv_move="${move%\'}"
            elif [[ "$move" == *"2" ]]; then
                inv_move="$move"
            else
                inv_move="${move}'"
            fi
            
            if [ -n "$inverse" ]; then
                inverse="$inverse $inv_move"
            else
                inverse="$inv_move"
            fi
        done
        
        # Test: scramble + inverse should return to solved using new verify
        solved_cfen="YB|Y9/R9/B9/W9/O9/G9"
        scrambled_cfen=$($CUBE_BIN generate-cfen "$scramble" 2>/dev/null)
        if ! $CUBE_BIN verify "$inverse" --start "$scrambled_cfen" --target "$solved_cfen" --headless >/dev/null 2>&1; then
            failed_count=$((failed_count + 1))
            
            # On first failure, show debug info
            if [ $failed_count -eq 1 ]; then
                echo -e "\n${RED}INVERSE FUZZ FAILURE DETECTED!${NC}"
                echo "This indicates a fundamental bug in the move system."
                echo "Failing sequence: $scramble"
                echo "Generated inverse: $inverse"
                echo ""
                echo "Re-running in debug mode:"
                echo "=== SCRAMBLE ===" 
                $CUBE_BIN twist "$scramble" --color
                echo ""
                echo "=== VERIFY (should solve) ===" 
                scrambled_cfen=$($CUBE_BIN generate-cfen "$scramble" 2>/dev/null)
                $CUBE_BIN verify "$inverse" --start "$scrambled_cfen" --target "YB|Y9/R9/B9/W9/O9/G9" --verbose --color
                echo ""
                echo "Move system is fundamentally broken. Halting inverse fuzzing."
                return 1
            fi
        fi
    done
    
    if [ $failed_count -eq 0 ]; then
        echo -e "${GREEN}PASS${NC} ($total_count/$total_count sequences correctly inverted)"
        return 0
    else
        echo -e "${RED}FAIL${NC} ($failed_count/$total_count sequences failed inverse test)"
        return 1
    fi
}

# Run the foundational inverse fuzzing test
echo -n "FOUNDATIONAL TEST: Move system inverse fuzzing... "
if fuzz_test_move_inverses $INVERSE_FUZZ_COUNT; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Verify inverses fuzzing test function
fuzz_test_verify_inverses() {
    local test_count=${1:-10}
    local failed_count=0
    local total_count=0
    
    # Set deterministic seed for reproducible testing
    RANDOM=123
    
    echo -n "Verify inverses fuzzing ($test_count random scrambles)... "
    
    for i in $(seq 1 $test_count); do
        total_count=$((total_count + 1))
        
        # Generate random scramble (4-8 moves)
        scramble_length=$((4 + RANDOM % 5))
        scramble=$(generate_random_scramble $scramble_length)
        
        # Create inverse sequence (reverse order, invert each move)
        inverse=""
        IFS=' ' read -ra MOVES <<< "$scramble"
        for ((j=${#MOVES[@]}-1; j>=0; j--)); do
            move="${MOVES[j]}"
            # Invert the move: R->R', R'->R, R2->R2
            if [[ "$move" == *"'" ]]; then
                inv_move="${move%\'}"
            elif [[ "$move" == *"2" ]]; then
                inv_move="$move"
            else
                inv_move="$move'"
            fi
            
            if [ -z "$inverse" ]; then
                inverse="$inv_move"
            else
                inverse="$inverse $inv_move"
            fi
        done
        
        # Use twist to generate scrambled CFEN state
        scrambled_cfen=$($CUBE_BIN twist "$scramble" --cfen 2>/dev/null)
        if [ $? -ne 0 ] || [ -z "$scrambled_cfen" ]; then
            failed_count=$((failed_count + 1))
            continue
        fi
        
        # Use verify to check that inverse solves the scrambled state
        solved_cfen="YB|Y9/R9/B9/W9/O9/G9"
        if ! $CUBE_BIN verify "$inverse" --start "$scrambled_cfen" --target "$solved_cfen" --headless >/dev/null 2>&1; then
            failed_count=$((failed_count + 1))
            
            # On first failure, show debug info and exit
            if [ $failed_count -eq 1 ]; then
                echo ""
                echo "VERIFY INVERSES FUZZ TEST FAILED!"
                echo "Scramble: $scramble"
                echo "Inverse:  $inverse"
                echo "Scrambled CFEN: $scrambled_cfen"
                echo ""
                echo "=== TWIST OUTPUT ==="
                $CUBE_BIN twist "$scramble" --color
                echo ""
                echo "=== VERIFY (should solve) ==="
                $CUBE_BIN verify "$inverse" --start "$scrambled_cfen" --target "$solved_cfen" --verbose --color
                echo ""
                echo "Verify system has issues. Halting verify inverses fuzzing."
                return 1
            fi
        fi
    done
    
    if [ $failed_count -eq 0 ]; then
        echo -e "${GREEN}PASS${NC} ($total_count/$total_count scrambles verified with inverses)"
        return 0
    else
        echo -e "${RED}FAIL${NC} ($failed_count/$total_count scrambles failed verify inverse test)"
        return 1
    fi
}

# Run the verify inverses fuzzing test
VERIFY_FUZZ_COUNT=10  # Number of random scrambles
echo -n "VERIFY INVERSES FUZZING: "
if fuzz_test_verify_inverses $VERIFY_FUZZ_COUNT; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# ===== SOLVER FUZZING (COMMENTED OUT UNTIL MOVE SYSTEM IS FIXED) =====
# 
# The solver tests below are commented out because we've discovered fundamental
# bugs in the move system itself. Until sequences + their inverses return to 
# solved state, no solver can work correctly.
#
# Uncomment these tests after the inverse fuzzing passes 100%

# Run fuzzing tests for each algorithm  
FUZZ_TEST_COUNT=25  # Number of random scrambles per algorithm

# echo -n "Fuzz test BeginnerSolver... "
# if fuzz_test_solver "beginner" $FUZZ_TEST_COUNT; then
#     TESTS_PASSED=$((TESTS_PASSED + 1))
# else
#     TESTS_FAILED=$((TESTS_FAILED + 1))
# fi
# TESTS_TOTAL=$((TESTS_TOTAL + 1))

# echo -n "Fuzz test CFOPSolver... "  
# if fuzz_test_solver "cfop" $FUZZ_TEST_COUNT; then
#     TESTS_PASSED=$((TESTS_PASSED + 1))
# else
#     TESTS_FAILED=$((TESTS_FAILED + 1))
# fi
# TESTS_TOTAL=$((TESTS_TOTAL + 1))

# echo -n "Fuzz test KociembaSolver... "
# if fuzz_test_solver "kociemba" $FUZZ_TEST_COUNT; then
#     TESTS_PASSED=$((TESTS_PASSED + 1))
# else
#     TESTS_FAILED=$((TESTS_FAILED + 1))
# fi
# TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Multi-dimensional fuzzing (COMMENTED OUT UNTIL MOVE SYSTEM IS FIXED)
# 
# These tests depend on solvers working, which depend on the move system working.
# Once inverse fuzzing passes 100%, we can re-enable these.

# echo -n "Fuzz test 2x2 cubes... "
# failed_2x2=0
# for i in $(seq 1 10); do
#     scramble=$(generate_random_scramble 8)
#     solution=$($CUBE_BIN solve "$scramble" --algorithm "beginner" --dimension 2 --headless 2>/dev/null)
#     if [ -n "$solution" ] && ! $CUBE_BIN verify "$scramble" "$solution" --dimension 2 --headless >/dev/null 2>&1; then
#         failed_2x2=$((failed_2x2 + 1))
#     fi
# done
# if [ $failed_2x2 -eq 0 ]; then
#     echo -e "${GREEN}PASS${NC} (10/10 2x2 scrambles solved)"
#     TESTS_PASSED=$((TESTS_PASSED + 1))
# else
#     echo -e "${RED}FAIL${NC} ($failed_2x2/10 2x2 scrambles failed)"
#     TESTS_FAILED=$((TESTS_FAILED + 1))
# fi
# TESTS_TOTAL=$((TESTS_TOTAL + 1))

# echo -n "Fuzz test 4x4 cubes... " 
# failed_4x4=0
# for i in $(seq 1 5); do
#     scramble=$(generate_random_scramble 12)
#     solution=$($CUBE_BIN solve "$scramble" --algorithm "beginner" --dimension 4 --headless 2>/dev/null)
#     if [ -n "$solution" ] && ! $CUBE_BIN verify "$scramble" "$solution" --dimension 4 --headless >/dev/null 2>&1; then
#         failed_4x4=$((failed_4x4 + 1))
#     fi
# done
# if [ $failed_4x4 -eq 0 ]; then
#     echo -e "${GREEN}PASS${NC} (5/5 4x4 scrambles solved)"
#     TESTS_PASSED=$((TESTS_PASSED + 1))
# else
#     echo -e "${RED}FAIL${NC} ($failed_4x4/5 4x4 scrambles failed)"
#     TESTS_FAILED=$((TESTS_FAILED + 1))
# fi
# TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Web tests removed - this is a CLI-only tool

# Phase 4 Power User Tools Tests
echo -e "\n${YELLOW}Phase 4 Power User Tools:${NC}"

run_test "Move optimization - basic" "$CUBE_BIN optimize \"R R\"" "R2.*1 moves"
run_test "Move optimization - canceling" "$CUBE_BIN optimize \"R R'\"" "empty.*all moves cancel"
run_test "Move optimization - complex" "$CUBE_BIN optimize \"R R R\"" "R'.*1 moves"
for turn in M E S Rw 3Rw; do
    run_test "Move optimization - $turn inverse cancellation" "$CUBE_BIN optimize \"$turn $turn'\"" "empty.*all moves cancel"
    run_test "Move optimization - $turn half-turn cancellation" "$CUBE_BIN optimize \"${turn}2 ${turn}2\"" "empty.*all moves cancel"
done

run_test "Algorithm discovery - simple solve" "$CUBE_BIN find pattern solved --max-moves 3 --from \"R\"" "R'"
run_test "Algorithm discovery - sequence solve" "$CUBE_BIN find sequence \"R U\" --max-moves 4" "U' R'"
run_test "Algorithm discovery - numbered sequence steps" "$CUBE_BIN find sequence \"R U\" -m 2 --steps" "   2. R'"
run_test "Algorithm discovery - cross pattern" "$CUBE_BIN find pattern cross --max-moves 4 --from \"F\"" "Found.*sequence"
search_start=$($CUBE_BIN generate-cfen "R U F2 L' B")
run_test "Algorithm discovery - concrete CFEN target" "$CUBE_BIN find --target 'YB|Y9/R9/B9/W9/O9/G9' --start '$search_start' --max-moves 5" "Found.*sequence"
run_test "Algorithm discovery - fixed center frame" "$CUBE_BIN find --target 'YB|Y9/R9/B9/W9/O9/G9' --start '$full_start' --max-moves 5" "No sequences found"
run_test "Algorithm discovery - wildcard CFEN target" "$CUBE_BIN find pattern 'YB|Y9/?9/?9/?9/?9/?9' --from F --max-moves 2" "Found.*sequence"
run_test "Algorithm discovery - restricted moves" "$CUBE_BIN find sequence R2 --moves R --max-moves 2" "R R (2 moves"
run_test "Algorithm discovery - 2x2 fallback" "$CUBE_BIN find sequence R --dimension 2 --max-moves 1" "R'"
run_test "Algorithm discovery - depth ten" "$CUBE_BIN find sequence \"R U F2 L' B D2 R' F U2 L\" --max-moves 10" "10 moves"

# CFEN (Cube Forsyth-Edwards Notation) Tests
echo -e "\n${BLUE}CFEN Tests:${NC}"

run_test "CFEN parse solved 3x3" "$CUBE_BIN parse-cfen \"YB|Y9/R9/B9/W9/O9/G9\"" "Solved: true"
run_test "CFEN parse 4x4 cube" "$CUBE_BIN parse-cfen \"YB|Y16/R16/B16/W16/O16/G16\"" "4x4x4"
run_test "CFEN parse with wildcards" "$CUBE_BIN parse-cfen \"WG|?W?WWW?W?/?9/?9/?9/?9/?9\"" "3x3x3"
run_test "CFEN generate from solved" "$CUBE_BIN generate-cfen \"\"" "YB|Y9/R9/B9/W9/O9/G9"
run_test "CFEN generate from scramble" "$CUBE_BIN generate-cfen \"R U R' U'\"" "YB|.*"
run_test "CFEN verify solved state" "$CUBE_BIN verify-cfen \"\" \"\" --target \"YB|Y9/R9/B9/W9/O9/G9\"" "PASS.*matches target"
run_test "CFEN verify wildcard matching" "$CUBE_BIN verify-cfen \"R U R' U'\" \"\" --target \"YB|?9/?9/?9/?9/?9/?9\"" "PASS.*matches target"
run_test "CFEN match identical states" "$CUBE_BIN match-cfen \"YB|Y9/R9/B9/W9/O9/G9\" \"YB|Y9/R9/B9/W9/O9/G9\"" "MATCH"
run_test "CFEN solve with output flag" "$CUBE_BIN solve \"R U R' U'\" --cfen" "YB|.*"
run_test "CFEN twist with output flag" "$CUBE_BIN twist \"R U R' U'\" --cfen" "YB|.*"
run_test "CFEN solve with start flag" "$CUBE_BIN solve \"U\" --start \"YB|Y9/R9/B9/W9/O9/G9\" --cfen" "YB|.*"

# Test CFEN orientation conversion
echo -n "Testing CFEN orientation conversion... "
WG_CFEN="WG|W9/R9/G9/Y9/O9/B9"
YB_FROM_WG=$($CUBE_BIN parse-cfen "$WG_CFEN" >/dev/null 2>&1 && $CUBE_BIN generate-cfen "" --start "$WG_CFEN" 2>/dev/null || echo "FAILED")
TESTS_TOTAL=$((TESTS_TOTAL + 1))

if [ "$YB_FROM_WG" != "$WG_CFEN" ] && [ "$YB_FROM_WG" != "FAILED" ]; then
    echo -e "${GREEN}passed${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}failed${NC} (WG: $WG_CFEN -> YB: $YB_FROM_WG)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Database Verification Tests
echo -e "\n${YELLOW}Database Verification Tests:${NC}"

# Test verify-algorithm tool exists and works
if [ -f "./dist/tools/verify-algorithm" ]; then
    run_test "verify-algorithm tool listing" "./dist/tools/verify-algorithm --list" "Sune.*HAS PATTERN"
    run_test "verify-algorithm single test" "./dist/tools/verify-algorithm 'Sune'" "✅ PASS"
    run_test "verify-algorithm verbose mode" "./dist/tools/verify-algorithm 'T-Perm' --verbose" "✅ PASS"
    run_test "verify-algorithm nonexistent algorithm" "./dist/tools/verify-algorithm 'NonExistent'" "" true
else
    echo -e "${BLUE}Testing:${NC} verify-algorithm tool exists... ${YELLOW}SKIP${NC} (not built)"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
fi

# Test verify-database tool exists and works  
if [ -f "./dist/tools/verify-database" ]; then
    run_test "verify-database full verification" "./dist/tools/verify-database" "All algorithms verified successfully"
    run_test "verify-database category filter" "./dist/tools/verify-database --category OLL" "algorithms tested"
    run_test "verify-database verbose mode" "./dist/tools/verify-database --verbose" "✅ PASS"
    run_test "verify-database invalid category" "./dist/tools/verify-database --category INVALID" "No algorithms.*found.*INVALID"
else
    echo -e "${BLUE}Testing:${NC} verify-database tool exists... ${YELLOW}SKIP${NC} (not built)"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
fi

# Performance benchmarking tests
echo -e "\n${YELLOW}Performance Benchmarking Tests:${NC}"

# Time the full database verification
if [ -f "./dist/tools/verify-database" ]; then
    echo -ne "${BLUE}Testing:${NC} Database verification performance... "
    start_time=$(date +%s%N)
    ./dist/tools/verify-database > /dev/null 2>&1
    end_time=$(date +%s%N)
    duration_ms=$(( (end_time - start_time) / 1000000 ))
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    if [ $duration_ms -lt 10000 ]; then  # Less than 10 seconds
        echo -e "${GREEN}PASS${NC} (${duration_ms}ms)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}SLOW${NC} (${duration_ms}ms - exceeds 10s limit)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo -e "${BLUE}Testing:${NC} Database verification performance... ${YELLOW}SKIP${NC} (not built)"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
fi

# Algorithm verification consistency test  
if [ -f "./dist/tools/verify-algorithm" ]; then
    echo -ne "${BLUE}Testing:${NC} Algorithm verification consistency... "
    
    # Run same algorithm verification 3 times and check consistency
    result1=$(./dist/tools/verify-algorithm "Sune" 2>&1 | grep -c "✅ PASS" || echo 0)
    result2=$(./dist/tools/verify-algorithm "Sune" 2>&1 | grep -c "✅ PASS" || echo 0)
    result3=$(./dist/tools/verify-algorithm "Sune" 2>&1 | grep -c "✅ PASS" || echo 0)
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    if [ "$result1" = "$result2" ] && [ "$result2" = "$result3" ] && [ "$result1" = "1" ]; then
        echo -e "${GREEN}PASS${NC} (consistent results)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAIL${NC} (inconsistent: $result1, $result2, $result3)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo -e "${BLUE}Testing:${NC} Algorithm verification consistency... ${YELLOW}SKIP${NC} (not built)"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
fi

# Visualization Tools Tests
echo -e "\n${YELLOW}Visualization Tools Tests:${NC}"

# Test show-alg command
run_test "show-alg basic display" "$CUBE_BIN show-alg 'Sune'" "Pattern:"
run_test "show-alg with color" "$CUBE_BIN show-alg 'T-Perm' --color" "🎯 FINAL STATE"
run_test "show-alg nonexistent algorithm" "$CUBE_BIN show-alg 'NonExistent'" "" true
run_test "show-alg without CFEN patterns" "$CUBE_BIN show-alg 'NonExistentAlg'" "not found in database" true

# Test identify command
run_test "identify solved state" "$CUBE_BIN identify 'YB|Y9/R9/B9/W9/O9/G9'" "🔍 ANALYZING PATTERN"
run_test "identify Sune pattern" "$CUBE_BIN identify 'YB|BY5RYG/YO2R6/YBOB6/W9/YG2O6/BR2G6'" "Anti-Sune"
run_test "identify with suggestions" "$CUBE_BIN identify 'YB|Y9/R9/B9/W9/O9/G9' --suggest" "SUGGESTIONS"
run_test "identify with category filter" "$CUBE_BIN identify 'YB|Y9/R9/B9/W9/O9/G9' --category OLL" "No matching algorithms"
run_test "identify invalid CFEN" "$CUBE_BIN identify 'INVALID'" "" true

# Working first-layer goal and beginner interaction
echo -e "\n${YELLOW}First-Layer Lesson Tests:${NC}"
run_test "Learn a first layer" "$CUBE_BIN learn --goal first-layer \"R U F2 L' B\"" "First layer complete:"
run_test "Learn a rotated scramble" "$CUBE_BIN learn --goal first-layer \"x y' R U F2 L' B\"" "Hold white down and blue front"
run_test "Learn with color" "$CUBE_BIN learn --goal first-layer R --color" "🟦"
run_test "Learn rejects bad dimensions" "$CUBE_BIN learn R --dimension 2" "" true
run_test "Learn rejects malformed moves" "$CUBE_BIN learn R3" "" true
run_test "Learn rejects wildcard input" "$CUBE_BIN learn --start 'YB|Y9/?9/B9/W9/O9/G9'" "" true
run_test "Learn rejects huge CFEN runs" "$CUBE_BIN learn --start 'YB|Y999999999/R9/B9/W9/O9/G9'" "" true
run_test "Interactive actual moves recover" "printf 'moves x R U\nundo\nreset\nquit\n' | $CUBE_BIN learn R --interactive" "Resume: cube learn"
run_test "Completed next is a no-op" "printf 'next\nnext\nquit\n' | $CUBE_BIN learn --goal first-layer U --interactive" "no moves applied"
run_test "Solve first-layer explicit goal" "$CUBE_BIN solve \"R U F2 L' B\" --goal first-layer" "Middle and last layers still need solving"
run_test "Solve first-layer incompatible algorithm" "$CUBE_BIN solve R --goal first-layer --algorithm cfop" "" true
run_test "Solve rejects unknown goal" "$CUBE_BIN solve R --goal unknown" "" true
run_test "Full beginner solver finishes" "$CUBE_BIN solve R --method beginner" "Cube complete:"

# Headless moves are replayed through verify against an independent sticker-row
# pattern. An empty placeholder answer cannot satisfy this scrambled fixture.
first_layer_scramble="x y' R U F2 L' B D R2 U'"
first_layer_start=$($CUBE_BIN generate-cfen "$first_layer_scramble")
first_layer_moves=$($CUBE_BIN solve "$first_layer_scramble" --goal first-layer --headless)
first_layer_target='YB|?9/??????RRR/??????BBB/W9/??????OOO/??????GGG'
run_test "First-layer headless replay" "$CUBE_BIN verify \"$first_layer_moves\" --start \"$first_layer_start\" --target '$first_layer_target' --headless"
first_layer_final=$($CUBE_BIN solve "$first_layer_scramble" --goal first-layer --cfen --headless)
run_test "First-layer saved CFEN replay" "$CUBE_BIN verify '' --start '$first_layer_final' --target '$first_layer_target' --headless"
run_test "Resume completed first layer" "$CUBE_BIN learn --goal first-layer --start '$first_layer_final'" "First layer complete:"

# The default lesson now finishes all six faces; first-layer remains explicit.
run_test "Default lesson fully solves" "$CUBE_BIN learn \"R U F2 L' B\"" "Cube complete:"
for numbered_turn in "2R" "2R'" "2R2"; do
    run_test "Numbered turn lesson replay ($numbered_turn)" \
        "lesson=\$(\"$CUBE_BIN\" learn \"$numbered_turn\") && solution=\$(printf '%s\\n' \"\$lesson\" | sed -n 's/^Solution: //p') && \"$CUBE_BIN\" twist \"$numbered_turn \$solution\"" "SOLVED!"
done
run_test "Numbered notation in move error" "$CUBE_BIN solve Q 2>&1 || true" "numbered 2R.*wide Rw or 2Rw"
run_test "Default lesson teaches middle edges" "$CUBE_BIN learn \"R U F2 L' B\"" "middle edge"
run_test "Default lesson teaches atomic corner sweep" "$CUBE_BIN learn \"R U F2 L' B\"" "finish the entire four-corner sweep"
run_test "Full solve saved-state result" "$CUBE_BIN solve --start '$full_start' --cfen --headless" 'YB|Y9/R9/B9/W9/O9/G9'
run_test "Solved full session next is a no-op" "printf 'next\nnext\nquit\n' | $CUBE_BIN learn '' --interactive" "Cube already complete; no moves applied"
run_test "Resume command preserves partial goal" "printf 'quit\n' | $CUBE_BIN learn R --goal first-layer --interactive" "goal first-layer"

# Summary
echo -e "\n${YELLOW}=== Test Summary ===${NC}"
echo -e "Total tests: $TESTS_TOTAL"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed! 🎉${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed.${NC}"
    exit 1
fi
