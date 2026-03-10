# Transpiled from English language source
# Test selective imports with commas
print("=== Testing Selective Imports with Commas ===")
# Import with commas
from math_library import square, cube, isEven

result = 0
result = square(3)
print("square(3) =", result)
result = cube(2)
print("cube(2) =", result)
check = False
check = isEven(10)
print("isEven(10) =", check)
print("=== Comma syntax works! ===")
