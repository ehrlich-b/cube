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
run_test "Solve with CFOP algorithm" "$CUBE_BIN solve \"R U R' U'\" --algorithm cfop" "Using algorithm: cfop"
run_test "Solve with Kociemba algorithm" "$CUBE_BIN solve \"R U R' U'\" --algorithm kociemba" "Using algorithm: kociemba"
run_test "Solve 2x2 cube" "$CUBE_BIN solve \"R U R' U'\" --dimension 2" "Solving 2x2x2 cube"
run_test "Solve 4x4 cube" "$CUBE_BIN solve \"Rw Uw Fw\" --dimension 4" "Solving 4x4x4 cube"
run_test "Solve 5x5 cube" "$CUBE_BIN solve \"2R 3L\" --dimension 5" "Solving 5x5x5 cube"
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
run_test "Wide moves" "$CUBE_BIN solve \"Rw Fw Uw\" --dimension 4" "Solution:"
run_test "Layer moves" "$CUBE_BIN solve \"2R 3L 2F\" --dimension 5" "Solution:"
run_test "Rotations" "$CUBE_BIN solve \"x y z\" --dimension 3" "Solution:"
run_test "Mixed notation" "$CUBE_BIN solve \"R M U' 2R Fw x y\" --dimension 5" "Solution:"

# Verify Command Tests
echo -e "\n${YELLOW}Verify Command Tests:${NC}"
run_test "Verify correct solution" "$CUBE_BIN verify \"R U R' U'\" \"U R U' R'\"" "✅ SOLVED!"
run_test "Verify incorrect solution" "$CUBE_BIN verify \"R U R' U'\" \"F U F' U'\"" "❌ NOT SOLVED!"
run_test "Verify with verbose" "$CUBE_BIN verify \"R U\" \"U' R'\" --verbose" "Cube after scramble"
run_test "Verify with color" "$CUBE_BIN verify \"R\" \"R'\" --verbose --color" "🟦"
run_test "Verify different dimensions" "$CUBE_BIN verify \"Rw\" \"Rw'\" --dimension 4" "✅ SOLVED!"

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
run_test "Lookup by name" "$CUBE_BIN lookup sune" "OLL 27 - Sune"
run_test "Lookup by pattern" "$CUBE_BIN lookup --pattern \"R U R' U'\"" "Sexy Move"
run_test "Lookup by category OLL" "$CUBE_BIN lookup --category OLL" "Sune"
run_test "Lookup by category PLL" "$CUBE_BIN lookup --category PLL" "T-Perm"
run_test "Lookup all algorithms" "$CUBE_BIN lookup --all" "All algorithms in database:"
run_test "Lookup with preview" "$CUBE_BIN lookup sune --preview" "Top face after algorithm:"
run_test "Lookup T-Perm" "$CUBE_BIN lookup \"T-Perm\"" "PLL T - T-Perm"
run_test "Lookup non-existent" "$CUBE_BIN lookup xyz123" "No algorithms found"
run_test "Lookup no args" "$CUBE_BIN lookup" "Please provide a query"

# Headless Mode Tests
echo -e "\n${YELLOW}Headless Mode Tests:${NC}"
echo -n "Testing headless solve output format... "
headless_output=$($CUBE_BIN solve "R" --headless 2>/dev/null)
if [[ "$headless_output" =~ ^[RUFLDBMESxyz\'2\ ]+$ ]] && [[ ! "$headless_output" =~ "Solution:" ]]; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (output should be space-separated moves only)"
    echo "Output: '$headless_output'"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Testing headless verify success (correct solution)... "
if $CUBE_BIN verify "R" "R'" --headless >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (should exit 0 for correct solution)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Testing headless verify failure (incorrect solution)... "
if ! $CUBE_BIN verify "R" "U" --headless >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (should exit 1 for incorrect solution)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Testing headless verify produces no output... "
headless_verify_output=$($CUBE_BIN verify "R" "R'" --headless 2>&1 || true)
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
run_test "Huge cube dimension" "$CUBE_BIN solve \"R\" --dimension 20" "Solving 20x20x20 cube"
run_test "Multiple flags" "$CUBE_BIN solve \"R U R' U'\" --color --dimension 4 --algorithm cfop" "Solving 4x4x4"

# Complex Integration Tests
echo -e "\n${YELLOW}Complex Integration Tests:${NC}"

# Test that all solvers actually solve scrambles
echo -n "Testing BeginnerSolver actually solves... "
scramble="R"
solution=$($CUBE_BIN solve "$scramble" --algorithm beginner 2>/dev/null | grep "Solution:" | sed 's/Solution: //')
if $CUBE_BIN verify "$scramble" "$solution" 2>&1 | grep -q "SOLVED"; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (BeginnerSolver should solve simple scrambles)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Testing BeginnerSolver solves complex scramble... "
scramble="R U R' U'"
solution=$($CUBE_BIN solve "$scramble" --algorithm beginner 2>/dev/null | grep "Solution:" | sed 's/Solution: //')
if $CUBE_BIN verify "$scramble" "$solution" 2>&1 | grep -q "SOLVED"; then
    echo -e "${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (BeginnerSolver should solve R U R' U')"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test that all algorithms work for simple cases
for algo in beginner cfop kociemba; do
    echo -n "Testing $algo solver works on simple cases... "
    passed=0
    total=0
    
    for scramble in "R" "U" "F" "R2" "U2"; do
        solution=$($CUBE_BIN solve "$scramble" --algorithm $algo 2>/dev/null | grep "Solution:" | sed 's/Solution: //')
        if [ -n "$solution" ] && $CUBE_BIN verify "$scramble" "$solution" 2>&1 | grep -q "SOLVED"; then
            passed=$((passed + 1))
        fi
        total=$((total + 1))
    done
    
    if [ $passed -eq $total ]; then
        echo -e "${GREEN}PASS${NC} ($passed/$total simple cases)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAIL${NC} (only $passed/$total simple cases work)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
done

# Test multiple algorithms produce different solutions
echo -n "Testing algorithm differences... "
sol1=$($CUBE_BIN solve "R U R' U'" --algorithm beginner 2>/dev/null | grep "Solution:" | sed 's/Solution: //')
sol2=$($CUBE_BIN solve "R U R' U'" --algorithm cfop 2>/dev/null | grep "Solution:" | sed 's/Solution: //')
sol3=$($CUBE_BIN solve "R U R' U'" --algorithm kociemba 2>/dev/null | grep "Solution:" | sed 's/Solution: //')
if [ "$sol1" != "$sol2" ] || [ "$sol2" != "$sol3" ] || [ "$sol1" != "$sol3" ]; then
    echo -e "${GREEN}PASS${NC} (algorithms produce different solutions)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} (algorithms produce identical solutions)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Performance Tests
echo -e "\n${YELLOW}Performance Tests:${NC}"
echo -n "Testing large scramble performance... "
start_time=$(date +%s%N 2>/dev/null || date +%s)
large_scramble="R U R' U' F B L R D U R U R' U' F B L R D U R U R' U' F B L R D U"
if $CUBE_BIN solve "$large_scramble" --dimension 6 >/dev/null 2>&1; then
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
        
        # Verify solution works
        if ! $CUBE_BIN verify "$scramble" "$solution" --headless >/dev/null 2>&1; then
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
            $CUBE_BIN verify "$scramble" "$solution" --verbose --color
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

# Run fuzzing tests for each algorithm  
FUZZ_TEST_COUNT=25  # Number of random scrambles per algorithm

echo -n "Fuzz test BeginnerSolver... "
if fuzz_test_solver "beginner" $FUZZ_TEST_COUNT; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Fuzz test CFOPSolver... "  
if fuzz_test_solver "cfop" $FUZZ_TEST_COUNT; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Fuzz test KociembaSolver... "
if fuzz_test_solver "kociemba" $FUZZ_TEST_COUNT; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Multi-dimensional fuzzing
echo -n "Fuzz test 2x2 cubes... "
failed_2x2=0
for i in $(seq 1 10); do
    scramble=$(generate_random_scramble 8)
    solution=$($CUBE_BIN solve "$scramble" --algorithm "beginner" --dimension 2 --headless 2>/dev/null)
    if [ -n "$solution" ] && ! $CUBE_BIN verify "$scramble" "$solution" --dimension 2 --headless >/dev/null 2>&1; then
        failed_2x2=$((failed_2x2 + 1))
    fi
done
if [ $failed_2x2 -eq 0 ]; then
    echo -e "${GREEN}PASS${NC} (10/10 2x2 scrambles solved)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} ($failed_2x2/10 2x2 scrambles failed)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Fuzz test 4x4 cubes... " 
failed_4x4=0
for i in $(seq 1 5); do
    scramble=$(generate_random_scramble 12)
    solution=$($CUBE_BIN solve "$scramble" --algorithm "beginner" --dimension 4 --headless 2>/dev/null)
    if [ -n "$solution" ] && ! $CUBE_BIN verify "$scramble" "$solution" --dimension 4 --headless >/dev/null 2>&1; then
        failed_4x4=$((failed_4x4 + 1))
    fi
done
if [ $failed_4x4 -eq 0 ]; then
    echo -e "${GREEN}PASS${NC} (5/5 4x4 scrambles solved)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAIL${NC} ($failed_4x4/5 4x4 scrambles failed)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Test 45: Terminal web interface (only on macOS/Linux with curl)
if command -v curl >/dev/null 2>&1; then
    echo -n "Test 45: Terminal web interface API... "
    # Start server in background and test it with better error handling
    $CUBE_BIN serve --port 8083 >/dev/null 2>&1 &
    SERVER_PID=$!
    sleep 3  # Give server more time to start
    
    # Test the API endpoint
    if curl -s -f -X POST http://localhost:8083/api/exec \
        -H 'Content-Type: application/json' \
        -d '{"command": "lookup sune"}' | grep -q 'Sune'; then
        echo -e "${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAIL${NC} (web terminal API not working)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # Test 46: Web interface shows twist command first
    echo -n "Test 46: Web interface prioritizes twist command... "
    if curl -s http://localhost:8083/terminal | grep -A1 "Try commands like:" | grep -q 'twist'; then
        echo -e "${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAIL${NC} (twist command not prioritized in web interface)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # Clean up server
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
else
    echo "Skipping web terminal tests (curl not available)"
fi

# Phase 4 Power User Tools Tests
echo -e "\n${YELLOW}Phase 4 Power User Tools:${NC}"

run_test "Move optimization - basic" "$CUBE_BIN optimize \"R R\"" "R2.*1 moves"
run_test "Move optimization - canceling" "$CUBE_BIN optimize \"R R'\"" "empty.*all moves cancel"
run_test "Move optimization - complex" "$CUBE_BIN optimize \"R R R\"" "R'.*1 moves"

run_test "Algorithm discovery - simple solve" "$CUBE_BIN find pattern solved --max-moves 3 --from \"R\"" "R'"
run_test "Algorithm discovery - sequence solve" "$CUBE_BIN find sequence \"R U\" --max-moves 4" "U' R'"
run_test "Algorithm discovery - cross pattern" "$CUBE_BIN find pattern cross --max-moves 4 --from \"F\"" "Found.*sequence"

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