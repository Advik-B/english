# Transpiled from English language source
# Test selective imports
print("=== Testing Selective Imports ===")
# Import only specific functions
from math_library import square, cube

result = 0
result = square(4)
print("square(4) =", result)
result = cube(2)
print("cube(2) =", result)
# Try to use isEven which should not be available
# This would cause an error if uncommented:
# Set result to the result of calling isEven with 4.
print("=== Selective import works! ===")
