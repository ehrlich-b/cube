#!/bin/bash
# Fuzz test for solvers - generates random scrambles and verifies solutions

CUBE_BIN="./dist/cube"
NUM_TESTS=20
MAX_SCRAMBLE_LENGTH=25  # Realistic scramble length for actual solving
TIMEOUT=120  # seconds per solve (2 minutes for complex scrambles)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if binary exists
if [ ! -f "$CUBE_BIN" ]; then
    echo -e "${RED}Error: $CUBE_BIN not found. Run 'make build' first.${NC}"
    exit 1
fi

# Possible moves for random scrambles
MOVES=("R" "R'" "R2" "L" "L'" "L2" "U" "U'" "U2" "D" "D'" "D2" "F" "F'" "F2" "B" "B'" "B2")

# Test counters
declare -A solver_pass_count
declare -A solver_fail_count
declare -A solver_timeout_count

solvers=("beginner" "kociemba" "cfop")

# Initialize counters
for solver in "${solvers[@]}"; do
    solver_pass_count[$solver]=0
    solver_fail_count[$solver]=0
    solver_timeout_count[$solver]=0
done

echo "=========================================="
echo "Rubik's Cube Solver Fuzz Testing"
echo "=========================================="
echo "Tests: $NUM_TESTS scrambles per solver"
echo "Scramble length: 20-25 moves (realistic difficulty)"
echo "Timeout: ${TIMEOUT}s per solve"
echo "Solvers: ${solvers[*]}"
echo "=========================================="
echo

for solver in "${solvers[@]}"; do
    echo -e "${YELLOW}Testing $solver solver...${NC}"

    for i in $(seq 1 $NUM_TESTS); do
        # Generate random scramble (20-25 moves for realistic difficulty)
        scramble_length=$((20 + RANDOM % 6))  # 20 to 25 moves
        scramble=""
        for j in $(seq 1 $scramble_length); do
            move_idx=$((RANDOM % ${#MOVES[@]}))
            if [ -n "$scramble" ]; then
                scramble="$scramble ${MOVES[$move_idx]}"
            else
                scramble="${MOVES[$move_idx]}"
            fi
        done

        echo -n "  Test $i/$NUM_TESTS: \"$scramble\" ... "

        # Solve with timeout
        solution_output=$(timeout $TIMEOUT $CUBE_BIN solve "$scramble" --algorithm "$solver" 2>&1 || echo "TIMEOUT")

        if echo "$solution_output" | grep -q "TIMEOUT"; then
            echo -e "${RED}TIMEOUT${NC}"
            ((solver_timeout_count[$solver]++))
            continue
        fi

        if echo "$solution_output" | grep -q "Error"; then
            echo -e "${RED}FAIL (solver error)${NC}"
            echo "     Output: $solution_output"
            ((solver_fail_count[$solver]++))
            continue
        fi

        # Extract solution from output
        solution=$(echo "$solution_output" | grep "^Solution:" | sed 's/Solution: //')

        # Empty solution is valid (cube was already solved)
        if ! echo "$solution_output" | grep -q "^Solution:"; then
            echo -e "${RED}FAIL (no solution in output)${NC}"
            echo "     Output: $solution_output"
            ((solver_fail_count[$solver]++))
            continue
        fi

        # Verify solution by applying scramble + solution and checking if solved
        combined="$scramble $solution"
        verify_output=$($CUBE_BIN twist "$combined" 2>&1 || echo "VERIFY_FAIL")

        if echo "$verify_output" | grep -q "✅ SOLVED!"; then
            echo -e "${GREEN}PASS${NC} ($solution)"
            ((solver_pass_count[$solver]++))
        else
            echo -e "${RED}FAIL (solution incorrect)${NC}"
            echo "     Scramble: $scramble"
            echo "     Solution: $solution"
            echo "     Output: $(echo "$verify_output" | tail -2)"
            ((solver_fail_count[$solver]++))
        fi
    done

    echo
done

# Print summary
echo "=========================================="
echo "SUMMARY"
echo "=========================================="

total_pass=0
total_fail=0
total_timeout=0

for solver in "${solvers[@]}"; do
    pass=${solver_pass_count[$solver]}
    fail=${solver_fail_count[$solver]}
    timeout=${solver_timeout_count[$solver]}
    total=$((pass + fail + timeout))

    total_pass=$((total_pass + pass))
    total_fail=$((total_fail + fail))
    total_timeout=$((total_timeout + timeout))

    echo
    echo "$solver solver:"
    echo "  Pass:    $pass/$NUM_TESTS ($(( pass * 100 / NUM_TESTS ))%)"
    echo "  Fail:    $fail/$NUM_TESTS"
    echo "  Timeout: $timeout/$NUM_TESTS"

    if [ $fail -gt 0 ] || [ $timeout -gt 0 ]; then
        echo -e "  Status:  ${RED}ISSUES DETECTED${NC}"
    else
        echo -e "  Status:  ${GREEN}ALL PASS${NC}"
    fi
done

echo
echo "=========================================="
echo "OVERALL:"
total_tests=$((${#solvers[@]} * NUM_TESTS))
echo "  Total tests: $total_tests"
echo "  Pass:        $total_pass ($(( total_pass * 100 / total_tests ))%)"
echo "  Fail:        $total_fail"
echo "  Timeout:     $total_timeout"

if [ $total_fail -gt 0 ] || [ $total_timeout -gt 0 ]; then
    echo -e "  Result:      ${RED}FAILURE${NC}"
    exit 1
else
    echo -e "  Result:      ${GREEN}SUCCESS${NC}"
    exit 0
fi
