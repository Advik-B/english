# Transpiled from English language source
import math

# Import Example
# Demonstrates importing code from other files
print("=== Import Example ===")
# Import math functions from another file
from math_library import *

print("Testing imported functions:")
# Use the imported square function
result = 0
result = square(5)
print("Square of 5:", result)
# Use the imported cube function
result = cube(3)
print("Cube of 3:", result)
# Use the imported isEven function
isEvenResult = False
isEvenResult = isEven(10)
print("Is 10 even?", isEvenResult)
isEvenResult = isEven(7)
print("Is 7 even?", isEvenResult)
# Use the imported constant
print("Value of pi:", math.pi)
print("=== Done! ===")
