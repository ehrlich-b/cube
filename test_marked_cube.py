#!/usr/bin/env python3

# Create a test case with unique markers for each sticker position
# This will reveal piece ordering bugs that are hidden with solid colors

def create_marked_cube_test():
    """
    Create test cases where each sticker has a unique identifier
    so we can track exact piece movements and detect inversions
    """
    
    # Instead of solid colors, use position-based markers
    # Top face: T00, T01, T02 / T10, T11, T12 / T20, T21, T22
    # This way inversions become visible even on single faces
    
    test_cases = []
    
    # Test each basic move on a "marked" cube
    basic_moves = ["R", "R'", "R2", "L", "L'", "L2", "U", "U'", "U2", 
                   "D", "D'", "D2", "F", "F'", "F2", "B", "B'", "B2"]
    
    for move in basic_moves:
        test_cases.append({
            "name": f"marked_cube_{move}",
            "move": move,
            "description": f"Apply {move} to marked cube to detect inversions"
        })
    
    return test_cases

# Test specific sequences that should be equivalent
def create_equivalence_tests():
    """
    Create tests for sequences that should produce identical results
    but use different move combinations
    """
    
    equivalence_tests = [
        # These should produce identical results
        ("R U R'", "R U R'"),  # Baseline
        ("R R R R", ""),       # 4x R should equal identity  
        ("R2 R2", ""),         # 2x R2 should equal identity
        ("R U' U R'", "R R'"), # U' U should cancel
        
        # More complex equivalences
        ("R U R' U'", "(R U R' U')"),  # Should be identical to itself
        
        # Commutator tests - [A,B] = A B A' B' should have specific properties
        ("R U R' U'", "U R U' R'"),  # Different orders, should be different results
    ]
    
    return equivalence_tests

if __name__ == "__main__":
    print("=== Marked Cube Tests ===")
    for test in create_marked_cube_test()[:5]:  # Show first 5
        print(f"Test: {test['name']}")
        print(f"Move: {test['move']}")
        print(f"Description: {test['description']}")
        print()
    
    print("=== Equivalence Tests ===") 
    for seq1, seq2 in create_equivalence_tests():
        print(f"'{seq1}' should equal '{seq2}'")
    print()
    
    print("Run these tests to find subtle inversion bugs!")