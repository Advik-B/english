# Transpiled from English language source
import math

# Test import everything
print("=== Testing Import Everything ===")
# Import all from the library
from math_library import *

result = 0
result = square(5)
print("square(5) =", result)
result = cube(3)
print("cube(3) =", result)
evenCheck = False
evenCheck = isEven(8)
print("isEven(8) =", evenCheck)
print("pi =", math.pi)
print("=== Import everything works! ===")
